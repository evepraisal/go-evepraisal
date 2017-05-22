package web

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/elazarl/go-bindata-assetfs"
	"github.com/evepraisal/go-evepraisal"
	"github.com/evepraisal/go-evepraisal/legacy"
	"github.com/husobee/vestigo"
	"github.com/mash/go-accesslog"
)

type MainPageStruct struct {
	Appraisal           *evepraisal.Appraisal
	TotalAppraisalCount int64
}

func (ctx *Context) HandleIndex(w http.ResponseWriter, r *http.Request) {
	txn := ctx.app.TransactionLogger.StartWebTransaction("view_index", w, r)
	defer txn.End()

	total, err := ctx.app.AppraisalDB.TotalAppraisals()
	if err != nil {
		ctx.renderErrorPage(w, http.StatusInternalServerError, "Something bad happened", err.Error())
		return
	}
	ctx.render(w, "main.html", MainPageStruct{TotalAppraisalCount: total})
}

func (ctx *Context) HandleAppraisal(w http.ResponseWriter, r *http.Request) {
	txn := ctx.app.TransactionLogger.StartWebTransaction("create_appraisal", w, r)
	defer txn.End()

	body := r.FormValue("raw_textarea")
	if len(body) > 200000 {
		ctx.renderErrorPage(w, http.StatusBadRequest, "Invalid input", "Input value is too big.")
		return
	}

	market := r.FormValue("market")
	marketID, err := strconv.ParseInt(market, 10, 64)
	if err == nil {
		var ok bool
		market, ok = legacy.MarketIDToName[marketID]
		if !ok {
			ctx.renderErrorPage(w, http.StatusBadRequest, "Invalid input", "Market not found.")
			return
		}
	}

	appraisal, err := ctx.app.StringToAppraisal(market, body)
	if err != nil {
		ctx.renderErrorPage(w, http.StatusBadRequest, "Invalid input", err.Error())
		return
	}

	err = ctx.app.AppraisalDB.PutNewAppraisal(appraisal)
	if err != nil {
		log.Printf("ERROR: saving appraisal: %s", err)
		ctx.renderErrorPage(w, http.StatusInternalServerError, "Something bad happened", err.Error())
		return
	}
	log.Printf("[New appraisal] id=%s, market=%s, items=%d, unparsed=%d", appraisal.ID, appraisal.MarketName, len(appraisal.Items), len(appraisal.Unparsed))

	err = ctx.render(w, "main.html", MainPageStruct{Appraisal: appraisal})
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
			ctx.renderErrorPage(w, http.StatusNotFound, "Not Found", "I couldn't find what you're looking for")
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
		ctx.renderErrorPage(w, http.StatusNotFound, "Not Found", "I couldn't find what you're looking for")
		return
	} else if err != nil {
		ctx.renderErrorPage(w, http.StatusInternalServerError, "Something bad happened", err.Error())
		return
	}

	sort.Slice(appraisal.Items, func(i, j int) bool {
		return appraisal.Items[i].SingleRepresentativePrice() > appraisal.Items[j].SingleRepresentativePrice()
	})

	ctx.render(w, "main.html", MainPageStruct{Appraisal: appraisal})
}

func (ctx *Context) HandleViewAppraisalJSON(w http.ResponseWriter, r *http.Request) {
	txn := ctx.app.TransactionLogger.StartWebTransaction("view_appraisal_json", w, r)
	defer txn.End()

	appraisalID := vestigo.Param(r, "appraisalID")
	appraisalID = strings.TrimSuffix(appraisalID, ".json")

	appraisal, err := ctx.app.AppraisalDB.GetAppraisal(appraisalID)
	if err == evepraisal.AppraisalNotFound {
		ctx.renderErrorPage(w, http.StatusNotFound, "Not Found", "I couldn't find what you're looking for")
		return
	} else if err != nil {
		ctx.renderErrorPage(w, http.StatusInternalServerError, "Something bad happened", err.Error())
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
		ctx.renderErrorPage(w, http.StatusNotFound, "Not Found", "I couldn't find what you're looking for")
		return
	} else if err != nil {
		ctx.renderErrorPage(w, http.StatusInternalServerError, "Something bad happened", err.Error())
		return
	}

	r.Header["Content-Type"] = []string{"application/text"}
	io.WriteString(w, appraisal.Raw)
}

func (ctx *Context) HandleLatestAppraisals(w http.ResponseWriter, r *http.Request) {
	txn := ctx.app.TransactionLogger.StartWebTransaction("view_latest_appraisals", w, r)
	defer txn.End()

	var limit int64
	var err error
	limit, err = strconv.ParseInt(r.FormValue("limit"), 10, 64)
	if err != nil {
		limit = 100
	}

	appraisals, err := ctx.app.AppraisalDB.LatestAppraisals(int(limit), r.FormValue("kind"))
	if err != nil {
		ctx.renderErrorPage(w, http.StatusInternalServerError, "Something bad happened", err.Error())
		return
	}

	ctx.render(w, "latest.html", struct{ Appraisals []evepraisal.Appraisal }{appraisals})
}

func (ctx *Context) HandleLegal(w http.ResponseWriter, r *http.Request) {
	txn := ctx.app.TransactionLogger.StartWebTransaction("view_legal", w, r)
	defer txn.End()
	ctx.render(w, "legal.html", "wat")
}

func (ctx *Context) HTTPHandler() http.Handler {
	router := vestigo.NewRouter()
	router.Get("/", ctx.HandleIndex)
	router.Post("/appraisal", ctx.HandleAppraisal)
	router.Post("/estimate", ctx.HandleAppraisal)
	router.Get("/a/:appraisalID", ctx.HandleViewAppraisal)
	router.Get("/e/:legacyAppraisalID", ctx.HandleViewAppraisal)
	router.Get("/latest", ctx.HandleLatestAppraisals)
	router.Get("/legal", ctx.HandleLegal)

	vestigo.CustomNotFoundHandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx.renderErrorPage(w, http.StatusNotFound, "Not Found", "I couldn't find what you're looking for")
		})

	vestigo.CustomMethodNotAllowedHandlerFunc(func(allowedMethods string) func(w http.ResponseWriter, r *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx.renderErrorPage(w, http.StatusInternalServerError, "Method not allowed", fmt.Sprintf("HTTP Method not allowed. What is allowed is: "+allowedMethods))
		}
	})

	mux := http.NewServeMux()

	// Route our bundled static files
	var fs = &assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo, Prefix: "/static/"}
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(fs)))

	// Mount our web app router to root
	mux.Handle("/", router)
	err := ctx.Reload()
	if err != nil {
		log.Fatal(err)
	}

	return accesslog.NewLoggingHandler(mux, accessLogger{})
}
