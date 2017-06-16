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

	"github.com/evepraisal/go-evepraisal"
	"github.com/evepraisal/go-evepraisal/legacy"
	"github.com/husobee/vestigo"
)

type AppraisalPage struct {
	Appraisal *evepraisal.Appraisal
	ShowFull  bool
}

func (ctx *Context) HandleAppraisal(w http.ResponseWriter, r *http.Request) {
	txn := ctx.App.TransactionLogger.StartWebTransaction("create_appraisal", w, r)
	defer txn.End()

	r.ParseMultipartForm(20 * 1000)

	var body string

	f, _, err := r.FormFile("uploadappraisal")

	if err == http.ErrMissingFile {
		body = r.FormValue("raw_textarea")
	} else if err != nil {
		ctx.renderErrorPage(r, w, http.StatusInternalServerError, "Something bad happened", err.Error())
		return
	} else {
		defer f.Close()
		bodyBytes, err := ioutil.ReadAll(f)
		if err != nil {
			ctx.renderErrorPage(r, w, http.StatusInternalServerError, "Something bad happened", err.Error())
			return
		}
		body = string(bodyBytes)
	}

	if len(body) > 200000 {
		ctx.renderErrorPage(r, w, http.StatusBadRequest, "Invalid input", "Input value is too big.")
		return
	}

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

	appraisal, err := ctx.App.StringToAppraisal(market, body)
	if err != nil {
		ctx.renderErrorPage(r, w, http.StatusBadRequest, "Invalid input", err.Error())
		return
	}

	appraisal.User = ctx.GetCurrentUser(r)

	err = ctx.App.AppraisalDB.PutNewAppraisal(appraisal)
	if err != nil {
		log.Printf("ERROR: saving appraisal: %s", err)
		ctx.renderErrorPage(r, w, http.StatusInternalServerError, "Something bad happened", err.Error())
		return
	}
	log.Printf("[New appraisal] id=%s, market=%s, items=%d, unparsed=%d", appraisal.ID, appraisal.MarketName, len(appraisal.Items), len(appraisal.Unparsed))

	// Set new session variable
	ctx.setDefaultMarket(r, w, market)

	sort.Slice(appraisal.Items, func(i, j int) bool {
		return appraisal.Items[i].RepresentativePrice() > appraisal.Items[j].RepresentativePrice()
	})

	ctx.render(r, w, "appraisal.html", AppraisalPage{Appraisal: appraisal})
}

func (ctx *Context) HandleViewAppraisal(w http.ResponseWriter, r *http.Request) {
	txn := ctx.App.TransactionLogger.StartWebTransaction("view_appraisal", w, r)
	defer txn.End()

	// Legacy Logic
	if vestigo.Param(r, "legacyAppraisalID") != "" {
		legacyAppraisalIDStr := vestigo.Param(r, "legacyAppraisalID")
		suffix := filepath.Ext(legacyAppraisalIDStr)
		legacyAppraisalIDStr = strings.TrimSuffix(legacyAppraisalIDStr, suffix)
		legacyAppraisalID, err := strconv.ParseUint(legacyAppraisalIDStr, 10, 64)
		if err != nil {
			ctx.renderErrorPage(r, w, http.StatusNotFound, "Not Found", "I couldn't find what you're looking for")
			return
		}
		vestigo.AddParam(r, "appraisalID", evepraisal.Uint64ToAppraisalID(legacyAppraisalID)+suffix)
	}

	appraisalID := vestigo.Param(r, "appraisalID")

	if strings.HasSuffix(appraisalID, ".json") {
		ctx.HandleViewAppraisalJSON(w, r)
		return
	}

	if strings.HasSuffix(appraisalID, ".raw") {
		ctx.HandleViewAppraisalRAW(w, r)
		return
	}

	appraisal, err := ctx.App.AppraisalDB.GetAppraisal(appraisalID)
	if err == evepraisal.AppraisalNotFound {
		ctx.renderErrorPage(r, w, http.StatusNotFound, "Not Found", "I couldn't find what you're looking for")
		return
	} else if err != nil {
		ctx.renderErrorPage(r, w, http.StatusInternalServerError, "Something bad happened", err.Error())
		return
	}

	sort.Slice(appraisal.Items, func(i, j int) bool {
		return appraisal.Items[i].RepresentativePrice() > appraisal.Items[j].RepresentativePrice()
	})

	ctx.render(r, w, "appraisal.html", AppraisalPage{Appraisal: appraisal, ShowFull: r.FormValue("full") != ""})
}

func (ctx *Context) HandleViewAppraisalJSON(w http.ResponseWriter, r *http.Request) {
	txn := ctx.App.TransactionLogger.StartWebTransaction("view_appraisal_json", w, r)
	defer txn.End()

	appraisalID := vestigo.Param(r, "appraisalID")
	appraisalID = strings.TrimSuffix(appraisalID, ".json")

	appraisal, err := ctx.App.AppraisalDB.GetAppraisal(appraisalID)
	if err == evepraisal.AppraisalNotFound {
		ctx.renderErrorPage(r, w, http.StatusNotFound, "Not Found", "I couldn't find what you're looking for")
		return
	} else if err != nil {
		ctx.renderErrorPage(r, w, http.StatusInternalServerError, "Something bad happened", err.Error())
		return
	}

	r.Header["Content-Type"] = []string{"application/json"}
	json.NewEncoder(w).Encode(appraisal)
}

func (ctx *Context) HandleViewAppraisalRAW(w http.ResponseWriter, r *http.Request) {
	txn := ctx.App.TransactionLogger.StartWebTransaction("view_appraisal_raw", w, r)
	defer txn.End()

	appraisalID := vestigo.Param(r, "appraisalID")
	appraisalID = strings.TrimSuffix(appraisalID, ".raw")

	appraisal, err := ctx.App.AppraisalDB.GetAppraisal(appraisalID)
	if err == evepraisal.AppraisalNotFound {
		ctx.renderErrorPage(r, w, http.StatusNotFound, "Not Found", "I couldn't find what you're looking for")
		return
	} else if err != nil {
		ctx.renderErrorPage(r, w, http.StatusInternalServerError, "Something bad happened", err.Error())
		return
	}

	r.Header["Content-Type"] = []string{"application/text"}
	io.WriteString(w, appraisal.Raw)
}
