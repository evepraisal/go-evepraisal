package web

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
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

type AppraisalPage struct {
	Appraisal *evepraisal.Appraisal
	ShowFull  bool
}

func (ctx *Context) HandleAppraisal(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(20 * 1000)

	market := r.FormValue("market")
	marketID, err := strconv.ParseInt(market, 10, 64)
	if err == nil {
		var ok bool
		market, ok = legacy.MarketIDToName[marketID]
		if !ok {
			ctx.renderErrorPage(r, w, http.StatusBadRequest, "Invalid input", "Market not found.")
			return
		}
	}

	var body string
	f, _, err := r.FormFile("uploadappraisal")
	if err == http.ErrNotMultipart || err == http.ErrMissingFile {
		body = r.FormValue("raw_textarea")
	} else if err != nil {
		ctx.renderServerError(r, w, err)
		return
	} else {
		defer f.Close()
		bodyBytes, err := ioutil.ReadAll(f)
		if err != nil {
			ctx.renderServerError(r, w, err)
			return
		}
		body = string(bodyBytes)
	}

	errorRoot := PageRoot{}
	errorRoot.UI.RawTextAreaDefault = body

	if len(body) > 200000 {
		ctx.renderErrorPageWithRoot(r, w, http.StatusBadRequest, "Invalid input", "Input value is too big.", errorRoot)
		return
	}

	if len(body) == 0 {
		ctx.renderErrorPageWithRoot(r, w, http.StatusBadRequest, "Invalid input", "Input value is empty.", errorRoot)
		return
	}

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

	err = ctx.App.AppraisalDB.PutNewAppraisal(appraisal)
	if err != nil {
		ctx.renderServerErrorWithRoot(r, w, err, errorRoot)
		return
	}

	username := ""
	user := ctx.GetCurrentUser(r)
	if user != nil {
		username = user.CharacterName
	}
	log.Printf("[New appraisal] id=%s, market=%s, items=%d, unparsed=%d, user=%s", appraisal.ID, appraisal.MarketName, len(appraisal.Items), len(appraisal.Unparsed), username)

	// Set new session variable
	ctx.setDefaultMarket(r, w, market)

	sort.Slice(appraisal.Items, func(i, j int) bool {
		return appraisal.Items[i].RepresentativePrice() > appraisal.Items[j].RepresentativePrice()
	})

	w.Header().Add("X-Appraisal-ID", appraisal.ID)

	ctx.render(r, w, "appraisal.html", AppraisalPage{Appraisal: appraisal})
}

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

	if strings.HasSuffix(appraisalID, ".json") {
		ctx.HandleViewAppraisalJSON(w, r, appraisalID)
		return
	}

	if strings.HasSuffix(appraisalID, ".raw") {
		ctx.HandleViewAppraisalRAW(w, r, appraisalID)
		return
	}

	appraisal, err := ctx.App.AppraisalDB.GetAppraisal(appraisalID)
	if err == evepraisal.AppraisalNotFound {
		ctx.renderErrorPage(r, w, http.StatusNotFound, "Not Found", "I couldn't find what you're looking for")
		return
	} else if err != nil {
		ctx.renderServerError(r, w, err)
		return
	}

	sort.Slice(appraisal.Items, func(i, j int) bool {
		return appraisal.Items[i].RepresentativePrice() > appraisal.Items[j].RepresentativePrice()
	})

	ctx.render(r, w, "appraisal.html", AppraisalPage{Appraisal: appraisal, ShowFull: r.FormValue("full") != ""})
}

func (ctx *Context) HandleViewAppraisalJSON(w http.ResponseWriter, r *http.Request, appraisalID string) {
	appraisalID = strings.TrimSuffix(appraisalID, ".json")

	appraisal, err := ctx.App.AppraisalDB.GetAppraisal(appraisalID)
	if err == evepraisal.AppraisalNotFound {
		ctx.renderErrorPage(r, w, http.StatusNotFound, "Not Found", "I couldn't find what you're looking for")
		return
	} else if err != nil {
		ctx.renderServerError(r, w, err)
		return
	}

	r.Header["Content-Type"] = []string{"application/json"}
	json.NewEncoder(w).Encode(appraisal)
}

func (ctx *Context) HandleViewAppraisalRAW(w http.ResponseWriter, r *http.Request, appraisalID string) {
	appraisalID = strings.TrimSuffix(appraisalID, ".raw")

	appraisal, err := ctx.App.AppraisalDB.GetAppraisal(appraisalID)
	if err == evepraisal.AppraisalNotFound {
		ctx.renderErrorPage(r, w, http.StatusNotFound, "Not Found", "I couldn't find what you're looking for")
		return
	} else if err != nil {
		ctx.renderServerError(r, w, err)
		return
	}

	r.Header["Content-Type"] = []string{"application/text"}
	io.WriteString(w, appraisal.Raw)
}
