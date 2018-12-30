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
	"net/url"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/evepraisal/go-evepraisal"
	"github.com/evepraisal/go-evepraisal/legacy"
	"github.com/evepraisal/go-evepraisal/parsers"
	"github.com/go-zoo/bone"
	"github.com/mssola/user_agent"
)

var (
	errInputTooBig = errors.New("Input value is too big")
	errInputEmpty  = errors.New("Input value is empty")

	formatJSON  = "json"
	formatRaw   = "raw"
	formatDebug = "debug"

	appraisalBodySizeLimit = int64(20 * 1000)
)

// AppraisalPage contains data used on the appraisal page
type AppraisalPage struct {
	Appraisal *evepraisal.Appraisal `json:"appraisal"`
	ShowFull  bool                  `json:"show_full,omitempty"`
	IsOwner   bool                  `json:"is_owner,omitempty"`
}

// AppraisalDebugPage is the data needed to render the
type AppraisalDebugPage struct {
	Appraisal    *evepraisal.Appraisal `json:"appraisal"`
	Lines        []AppraisalDebugLine  `json:"lines"`
	ParserResult parsers.ParserResult  `json:"parser_result"`
}

// AppraisalDebugLine represents a single line of an appraisal along with how it was parsed
type AppraisalDebugLine struct {
	Number int    `json:"number"`
	Parsed bool   `json:"parsed"`
	Parser string `json:"parser"`
	Text   string `json:"text"`
}

func makeAppraisalURL(appraisal *evepraisal.Appraisal) *url.URL {
	u := &url.URL{}
	if appraisal.Private {
		u.Path = fmt.Sprintf("/a/%s/%s", appraisal.ID, appraisal.PrivateToken)
	} else {
		u.Path = fmt.Sprintf("/a/%s", appraisal.ID)
	}
	return u
}

func maybeAddLiveParam(appraisal *evepraisal.Appraisal, u *url.URL) {
	q := u.Query()
	if appraisal.Live {
		q.Set("live", "yes")
	}
	u.RawQuery = q.Encode()
}

func appraisalLink(appraisal *evepraisal.Appraisal) string {
	u := makeAppraisalURL(appraisal)
	maybeAddLiveParam(appraisal, u)
	return u.String()
}

func liveAppraisalLink(appraisal *evepraisal.Appraisal) string {
	u := makeAppraisalURL(appraisal)
	q := u.Query()
	q.Set("live", "yes")
	u.RawQuery = q.Encode()
	return u.String()
}

func normalAppraisalLink(appraisal *evepraisal.Appraisal) string {
	u := makeAppraisalURL(appraisal)
	return u.String()
}

func rawAppraisalLink(appraisal *evepraisal.Appraisal) string {
	u := makeAppraisalURL(appraisal)
	u.Path = u.Path + ".raw"
	maybeAddLiveParam(appraisal, u)
	return u.String()
}

func jsonAppraisalLink(appraisal *evepraisal.Appraisal) string {
	u := makeAppraisalURL(appraisal)
	u.Path = u.Path + ".json"
	maybeAddLiveParam(appraisal, u)
	return u.String()
}

func isMultiPart(r *http.Request) bool {
	return strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data")
}

func isURLEncodedFormData(r *http.Request) bool {
	return r.Header.Get("Content-Type") == "application/x-www-form-urlencoded"
}

func getRequestParam(r *http.Request, name string) string {
	if isMultiPart(r) || isURLEncodedFormData(r) {
		v := r.FormValue(name)
		if v != "" {
			return v
		}
		return r.URL.Query().Get(name)
	}
	return r.URL.Query().Get(name)
}

func parseAppraisalBody(r *http.Request) (string, error) {
	// Parse body
	var (
		f    io.ReadCloser
		err  error
		body string
	)

	if isMultiPart(r) || isURLEncodedFormData(r) {
		r.ParseMultipartForm(appraisalBodySizeLimit)
		f, _, err = r.FormFile("uploadappraisal")
		if err != nil && err != http.ErrNotMultipart && err != http.ErrMissingFile {
			return "", err
		}
	} else {
		f = r.Body
	}

	if f != nil {
		bodyBytes, err := ioutil.ReadAll(io.LimitReader(f, appraisalBodySizeLimit))
		if err != nil {
			return "", err
		}
		body = string(bodyBytes)
		defer f.Close()
	}

	if body == "" {
		body = getRequestParam(r, "raw_textarea")
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

	body, err := parseAppraisalBody(r)
	if err != nil {
		ctx.renderErrorPage(r, w, http.StatusBadRequest, "Invalid input", err.Error())
		return
	}

	persist := getRequestParam(r, "persist") != "no"
	pricePercentageStr := "100"
	if getRequestParam(r, "price_percentage") != "" {
		pricePercentageStr = getRequestParam(r, "price_percentage")
	}
	pricePercentage, err := strconv.ParseFloat(pricePercentageStr, 64)
	if err != nil || pricePercentage <= 0 || pricePercentage > 1000 {
		ctx.renderErrorPage(r, w, http.StatusBadRequest, "Invalid price_percentage value", err.Error())
		return
	}

	expireAfterStr := getRequestParam(r, "expire_after")
	if expireAfterStr == "" {
		expireAfterStr = "360h"
	}
	expireAfter, err := time.ParseDuration(expireAfterStr)
	if err != nil {
		ctx.renderErrorPage(r, w, http.StatusBadRequest, "Invalid expire_after value", err.Error())
		return
	}

	if expireAfter < time.Minute || expireAfter > 90*24*time.Hour {
		ctx.renderErrorPage(r, w, http.StatusBadRequest, "Invalid expire_after value.", "It needs to be between 1m and 2160h")
		return
	}

	root := PageRoot{}
	root.UI.RawTextAreaDefault = body

	// Parse Market
	market := getRequestParam(r, "market")

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
		ctx.renderErrorPageWithRoot(r, w, http.StatusBadRequest, "Invalid input", "A market is required.", root)
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
		ctx.renderErrorPageWithRoot(r, w, http.StatusBadRequest, "Invalid input", "Given market is not valid.", root)
		return
	}

	user := ctx.GetCurrentUser(r)

	visibility := getRequestParam(r, "visibility")
	private := false
	if visibility == "private" && user != nil {
		private = true
	}

	// Actually do the appraisal
	appraisal, err := ctx.App.StringToAppraisal(market, body, pricePercentage)
	if err == evepraisal.ErrNoValidLinesFound {
		log.Println("No valid lines found:", spew.Sdump(body))
		ctx.renderErrorPageWithRoot(r, w, http.StatusBadRequest, "Invalid input", err.Error(), root)
		return
	} else if err != nil {
		ctx.renderErrorPageWithRoot(r, w, http.StatusBadRequest, "Invalid input", err.Error(), root)
		return
	}

	appraisal.User = ctx.GetCurrentUser(r)
	appraisal.Private = private
	if private {
		appraisal.PrivateToken = NewPrivateAppraisalToken()
	}
	appraisal.ExpireMinutes = int64(expireAfter.Minutes())

	// Persist Appraisal to the database
	if persist {
		err = ctx.App.AppraisalDB.PutNewAppraisal(appraisal)
		if err != nil {
			ctx.renderServerErrorWithRoot(r, w, err, root)
			return
		}
	} else {
		go ctx.App.AppraisalDB.IncrementTotalAppraisals()
	}

	// Log for later analyics
	log.Println(appraisal.Summary())

	// Set new session variable
	ctx.setSessionValue(r, w, "market", market)
	ctx.setSessionValue(r, w, "visibility", visibility)
	ctx.setSessionValue(r, w, "persist", persist)
	ctx.setSessionValue(r, w, "price_percentage", pricePercentage)
	ctx.setSessionValue(r, w, "expire_after", expireAfterStr)

	sort.Slice(appraisal.Items, func(i, j int) bool {
		return appraisal.Items[i].RepresentativePrice() > appraisal.Items[j].RepresentativePrice()
	})

	// Render the new appraisal to the screen (there is no redirect here, we set the URL using javascript later)
	w.Header().Add("X-Appraisal-ID", appraisal.ID)
	root.Page = AppraisalPage{
		IsOwner:   IsAppraisalOwner(user, appraisal),
		Appraisal: cleanAppraisal(appraisal),
	}
	ctx.renderWithRoot(r, w, "appraisal.html", root)
}

// HandleAppraisalStructured is the handler for POST /appraisal/structured.json
func (ctx *Context) HandleAppraisalStructured(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("format") != formatJSON {
		ctx.renderErrorPage(r, w, http.StatusNotFound, "Not found", "/appraisal/structured is only available for the JSON format.")
		return
	}

	decoder := json.NewDecoder(r.Body)
	var spec = struct {
		MarketName string `json:"market_name"`
		Items      []struct {
			TypeID   int64  `json:"type_id"`
			Name     string `json:"name"`
			Quantity int64  `json:"quantity"`
		} `json:"items"`
	}{}
	err := decoder.Decode(&spec)
	if err != nil {
		ctx.renderErrorPage(r, w, http.StatusBadRequest, "Invalid input", err.Error())
		return
	}

	if len(spec.Items) == 0 {
		ctx.renderErrorPage(r, w, http.StatusBadRequest, "Invalid input", "No 'items' given.")
		return
	}

	// Invalid market given
	foundMarket := false
	for _, m := range selectableMarkets {
		if m.Name == spec.MarketName {
			foundMarket = true
			break
		}
	}
	if !foundMarket {
		ctx.renderErrorPage(r, w, http.StatusBadRequest, "Invalid input", "Given market is not valid.")
		return
	}

	appraisal := &evepraisal.Appraisal{
		Created:    time.Now().Unix(),
		Kind:       "structured",
		Items:      make([]evepraisal.AppraisalItem, len(spec.Items)),
		MarketName: spec.MarketName,
	}

	for i, item := range spec.Items {
		if item.Name == "" && item.TypeID == 0 {
			ctx.renderErrorPage(r, w, http.StatusBadRequest, "Invalid input", fmt.Sprintf("Item at index %d does not have a 'name' or 'type_id'", i))
			return
		}

		if item.Quantity == 0 {
			item.Quantity = 1
		}
		appraisal.Items[i].Name = item.Name
		appraisal.Items[i].TypeID = item.TypeID
		appraisal.Items[i].Quantity = item.Quantity
	}

	ctx.App.PopulateItems(appraisal)

	go ctx.App.AppraisalDB.IncrementTotalAppraisals()

	// Log for later analyics
	log.Println(appraisal.Summary())

	sort.Slice(appraisal.Items, func(i, j int) bool {
		return appraisal.Items[i].RepresentativePrice() > appraisal.Items[j].RepresentativePrice()
	})

	ctx.render(r, w, "appraisal.html",
		AppraisalPage{
			Appraisal: cleanAppraisal(appraisal),
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

	appraisal, err := ctx.App.AppraisalDB.GetAppraisal(appraisalID, !user_agent.New(r.UserAgent()).Bot())
	if err == evepraisal.ErrAppraisalNotFound {
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

	appraisal.Live = getRequestParam(r, "live") == "yes"
	if appraisal.Live {
		ctx.App.PopulateItems(appraisal)
		appraisal.Created = time.Now().Unix()
	}

	appraisal = cleanAppraisal(appraisal)

	sort.Slice(appraisal.Items, func(i, j int) bool {
		return appraisal.Items[i].RepresentativePrice() > appraisal.Items[j].RepresentativePrice()
	})

	if r.Header.Get("format") == formatJSON {
		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(appraisal)
		return
	}

	if r.Header.Get("format") == formatRaw {
		io.WriteString(w, appraisal.Raw)
		return
	}

	if r.Header.Get("format") == formatDebug {
		ctx.renderAppraisalDebug(w, r, appraisal)
		return
	}

	ctx.render(r, w, "appraisal.html",
		AppraisalPage{
			Appraisal: appraisal,
			ShowFull:  getRequestParam(r, "full") != "",
			IsOwner:   isOwner,
		})
}

func (ctx *Context) renderAppraisalDebug(w http.ResponseWriter, r *http.Request, appraisal *evepraisal.Appraisal) {

	lines := strings.Split(appraisal.Raw, "\n")
	debugLines := make([]AppraisalDebugLine, len(lines))

	lineParsers := make(map[int]string)
	for parser, lines := range appraisal.ParserLines {
		for _, line := range lines {
			lineParsers[line] = parser
		}
	}

	for i, line := range lines {
		_, unparsed := appraisal.Unparsed[i]
		parser, ok := lineParsers[i]
		if !ok {
			parser = "unknown"
		}

		debugLines[i] = AppraisalDebugLine{
			Number: i,
			Parsed: !unparsed,
			Parser: parser,
			Text:   line,
		}
	}

	result, _ := ctx.App.Parser(parsers.StringToInput(appraisal.Raw))
	ctx.render(r, w, "appraisal_debug.html",
		AppraisalDebugPage{
			Appraisal:    appraisal,
			Lines:        debugLines,
			ParserResult: result,
		})
}

// HandleDeleteAppraisal is the handler for POST /a/delete/[id]
func (ctx *Context) HandleDeleteAppraisal(w http.ResponseWriter, r *http.Request) {
	appraisalID := bone.GetValue(r, "appraisalID")
	appraisal, err := ctx.App.AppraisalDB.GetAppraisal(appraisalID, false)
	if err == evepraisal.ErrAppraisalNotFound {
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
	if err == evepraisal.ErrAppraisalNotFound {
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
