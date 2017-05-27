package web

import (
	"encoding/json"
	"io"
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
}

func (ctx *Context) HandleAppraisal(w http.ResponseWriter, r *http.Request) {
	txn := ctx.app.TransactionLogger.StartWebTransaction("create_appraisal", w, r)
	defer txn.End()

	body := r.FormValue("raw_textarea")
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

	appraisal, err := ctx.app.StringToAppraisal(market, body)
	if err != nil {
		ctx.renderErrorPage(r, w, http.StatusBadRequest, "Invalid input", err.Error())
		return
	}

	err = ctx.app.AppraisalDB.PutNewAppraisal(appraisal)
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

	err = ctx.render(r, w, "appraisal.html", AppraisalPage{Appraisal: appraisal})
}

func (ctx *Context) HandleViewAppraisal(w http.ResponseWriter, r *http.Request) {
	txn := ctx.app.TransactionLogger.StartWebTransaction("view_appraisal", w, r)
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

	appraisal, err := ctx.app.AppraisalDB.GetAppraisal(appraisalID)
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

	ctx.render(r, w, "appraisal.html", AppraisalPage{Appraisal: appraisal})
}

func (ctx *Context) HandleViewAppraisalJSON(w http.ResponseWriter, r *http.Request) {
	txn := ctx.app.TransactionLogger.StartWebTransaction("view_appraisal_json", w, r)
	defer txn.End()

	appraisalID := vestigo.Param(r, "appraisalID")
	appraisalID = strings.TrimSuffix(appraisalID, ".json")

	appraisal, err := ctx.app.AppraisalDB.GetAppraisal(appraisalID)
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
	txn := ctx.app.TransactionLogger.StartWebTransaction("view_appraisal_raw", w, r)
	defer txn.End()

	appraisalID := vestigo.Param(r, "appraisalID")
	appraisalID = strings.TrimSuffix(appraisalID, ".raw")

	appraisal, err := ctx.app.AppraisalDB.GetAppraisal(appraisalID)
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
