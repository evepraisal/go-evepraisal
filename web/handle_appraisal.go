package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/evepraisal/go-evepraisal"
	"github.com/evepraisal/go-evepraisal/legacy"
	"github.com/go-zoo/bone"
)

var (
	errInputTooBig = errors.New("Input value is too big")
	errInputEmpty  = errors.New("Input value is empty")
)

// AppraisalPage contains data used on the appraisal page
type AppraisalPage struct {
	Appraisal *evepraisal.Appraisal `json:"appraisal"`
	ShowFull  bool                  `json:"show_full,omitempty"`
	IsOwner   bool                  `json:"is_owner,omitempty"`
}

func appraisalLink(appraisal *evepraisal.Appraisal) string {
	if appraisal.Private {
		return fmt.Sprintf("/a/%s/%s", appraisal.ID, appraisal.PrivateToken)
	}
	return fmt.Sprintf("/a/%s", appraisal.ID)
}

func parseAppraisalBody(r *http.Request) (string, error) {
	// Parse body
	r.ParseMultipartForm(20 * 1000)

	var body string
	f, _, err := r.FormFile("uploadappraisal")
	if err == http.ErrNotMultipart || err == http.ErrMissingFile {
		body = r.FormValue("raw_textarea")
	} else if err != nil {
		return "", err
	} else {
		defer f.Close()
		bodyBytes, err := ioutil.ReadAll(f)
		if err != nil {
			return "", err
		}
		body = string(bodyBytes)
	}
	if len(body) > 200000 {
		return "", errInputTooBig
	}

	if len(body) == 0 {
		return "", errInputEmpty
	}
	return body, nil
}

// HandleAppraisal is the handler for POST /appraisal
func (ctx *Context) HandleAppraisal(w http.ResponseWriter, r *http.Request) {

	persist := r.FormValue("persist") != "no"

	body, err := parseAppraisalBody(r)
	if err != nil {
		ctx.renderErrorPage(r, w, http.StatusBadRequest, "Invalid input", err.Error())
		return
	}

	errorRoot := PageRoot{}
	errorRoot.UI.RawTextAreaDefault = body

	// Parse Market
	market := r.FormValue("market")

	// Legacy Market ID
	marketID, err := strconv.ParseInt(market, 10, 64)
	if err == nil {
		var ok bool
		market, ok = legacy.MarketIDToName[marketID]
		if !ok {
			ctx.renderErrorPage(r, w, http.StatusBadRequest, "Invalid input", "Market not found.")
			return
		}
	}

	// No market given
	if market == "" {
		ctx.renderErrorPageWithRoot(r, w, http.StatusBadRequest, "Invalid input", "A market is required.", errorRoot)
		return
	}

	// Invalid market given
	foundMarket := false
	for _, m := range selectableMarkets {
		if m.Name == market {
			foundMarket = true
			break
		}
	}
	if !foundMarket {
		ctx.renderErrorPageWithRoot(r, w, http.StatusBadRequest, "Invalid input", "Given market is not valid.", errorRoot)
		return
	}

	user := ctx.GetCurrentUser(r)

	visibility := r.FormValue("visibility")
	private := false
	if visibility == "private" && user != nil {
		private = true
	}

	// Actually do the appraisal
	appraisal, err := ctx.App.StringToAppraisal(market, body)
	if err == evepraisal.ErrNoValidLinesFound {
		log.Println("No valid lines found:", spew.Sdump(body))
		ctx.renderErrorPageWithRoot(r, w, http.StatusBadRequest, "Invalid input", err.Error(), errorRoot)
		return
	} else if err != nil {
		ctx.renderErrorPageWithRoot(r, w, http.StatusBadRequest, "Invalid input", err.Error(), errorRoot)
		return
	}

	appraisal.User = ctx.GetCurrentUser(r)
	appraisal.Private = private
	appraisal.PrivateToken = NewPrivateAppraisalToken()

	// Persist Appraisal to the database
	if persist {
		err = ctx.App.AppraisalDB.PutNewAppraisal(appraisal)
		if err != nil {
			ctx.renderServerErrorWithRoot(r, w, err, errorRoot)
			return
		}
	}

	username := ""
	if user != nil {
		username = user.CharacterName
	}
	log.Printf("[New appraisal] id=%s, market=%s, kind=%s, items=%d, unparsed=%d, user=%s", appraisal.ID, appraisal.MarketName, appraisal.Kind, len(appraisal.Items), len(appraisal.Unparsed), username)

	// Set new session variable
	ctx.setSessionValue(r, w, "market", market)
	ctx.setSessionValue(r, w, "visibility", visibility)
	ctx.setSessionValue(r, w, "persist", persist)

	sort.Slice(appraisal.Items, func(i, j int) bool {
		return appraisal.Items[i].RepresentativePrice() > appraisal.Items[j].RepresentativePrice()
	})

	// Render the new appraisal to the screen (there is no redirect here, we set the URL using javascript later)
	w.Header().Add("X-Appraisal-ID", appraisal.ID)
	ctx.render(r, w, "appraisal.html",
		AppraisalPage{
			Appraisal: cleanAppraisal(appraisal),
			IsOwner:   IsAppraisalOwner(user, appraisal),
		},
	)
}

// HandleViewAppraisal is the handler for /a/[id]
func (ctx *Context) HandleViewAppraisal(w http.ResponseWriter, r *http.Request) {
	// Legacy Logic
	appraisalID := bone.GetValue(r, "appraisalID")
	if bone.GetValue(r, "legacyAppraisalID") != "" {
		legacyAppraisalIDStr := bone.GetValue(r, "legacyAppraisalID")
		suffix := filepath.Ext(legacyAppraisalIDStr)
		legacyAppraisalIDStr = strings.TrimSuffix(legacyAppraisalIDStr, suffix)
		legacyAppraisalID, err := strconv.ParseUint(legacyAppraisalIDStr, 10, 64)
		if err != nil {
			ctx.renderErrorPage(r, w, http.StatusNotFound, "Not Found", "I couldn't find what you're looking for")
			return
		}
		appraisalID = evepraisal.Uint64ToAppraisalID(legacyAppraisalID) + suffix
	}

	appraisal, err := ctx.App.AppraisalDB.GetAppraisal(appraisalID)
	if err == evepraisal.AppraisalNotFound {
		ctx.renderErrorPage(r, w, http.StatusNotFound, "Not Found", "I couldn't find what you're looking for")
		return
	} else if err != nil {
		ctx.renderServerError(r, w, err)
		return
	}

	user := ctx.GetCurrentUser(r)
	isOwner := IsAppraisalOwner(user, appraisal)

	if appraisal.Private {
		correctToken := appraisal.PrivateToken == bone.GetValue(r, "privateToken")
		if !(isOwner || correctToken) {
			ctx.renderErrorPage(r, w, http.StatusNotFound, "Not Found", "I couldn't find what you're looking for")
			return
		}
	} else if bone.GetValue(r, "privateToken") != "" {
		ctx.renderErrorPage(r, w, http.StatusNotFound, "Not Found", "I couldn't find what you're looking for")
		return
	}

	appraisal = cleanAppraisal(appraisal)

	sort.Slice(appraisal.Items, func(i, j int) bool {
		return appraisal.Items[i].RepresentativePrice() > appraisal.Items[j].RepresentativePrice()
	})

	if r.Header.Get("format") == "json" {
		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(appraisal)
		return
	}

	if r.Header.Get("format") == "raw" {
		io.WriteString(w, appraisal.Raw)
		return
	}

	ctx.render(r, w, "appraisal.html",
		AppraisalPage{
			Appraisal: appraisal,
			ShowFull:  r.FormValue("full") != "",
			IsOwner:   isOwner,
		})
}

// HandleDeleteAppraisal is the handler for POST /a/delete/[id]
func (ctx *Context) HandleDeleteAppraisal(w http.ResponseWriter, r *http.Request) {
	appraisalID := bone.GetValue(r, "appraisalID")
	appraisal, err := ctx.App.AppraisalDB.GetAppraisal(appraisalID)
	if err == evepraisal.AppraisalNotFound {
		ctx.renderErrorPage(r, w, http.StatusNotFound, "Not Found", "I couldn't find what you're looking for")
		return
	} else if err != nil {
		ctx.renderServerError(r, w, err)
		return
	}

	if !IsAppraisalOwner(ctx.GetCurrentUser(r), appraisal) {
		ctx.renderErrorPage(r, w, http.StatusNotFound, "Not Found", "I couldn't find what you're looking for")
		return
	}

	err = ctx.App.AppraisalDB.DeleteAppraisal(appraisalID)
	if err == evepraisal.AppraisalNotFound {
		ctx.renderErrorPage(r, w, http.StatusNotFound, "Not Found", "I couldn't find what you're looking for")
		return
	} else if err != nil {
		ctx.renderServerError(r, w, err)
		return
	}

	ctx.setFlashMessage(r, w, FlashMessage{Message: fmt.Sprintf("Appraisal %s has been deleted.", appraisalID), Severity: "success"})
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// NewPrivateAppraisalToken returns a new token to use for private appraisals
func NewPrivateAppraisalToken() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, 16)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}
