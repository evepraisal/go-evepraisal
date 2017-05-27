package web

import (
	"fmt"
	"log"
	"net/http"

	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/context"
	"github.com/husobee/vestigo"
	"github.com/mash/go-accesslog"
)

func (ctx *Context) HandleIndex(w http.ResponseWriter, r *http.Request) {
	txn := ctx.app.TransactionLogger.StartWebTransaction("view_index", w, r)
	defer txn.End()

	total, err := ctx.app.AppraisalDB.TotalAppraisals()
	if err != nil {
		ctx.renderErrorPage(r, w, http.StatusInternalServerError, "Something bad happened", err.Error())
		return
	}
	ctx.render(r, w, "main.html", struct{ TotalAppraisalCount int64 }{TotalAppraisalCount: total})
}

func (ctx *Context) HandleLegal(w http.ResponseWriter, r *http.Request) {
	txn := ctx.app.TransactionLogger.StartWebTransaction("view_legal", w, r)
	defer txn.End()
	ctx.render(r, w, "legal.html", nil)
}

func (ctx *Context) HandleHelp(w http.ResponseWriter, r *http.Request) {
	txn := ctx.app.TransactionLogger.StartWebTransaction("view_help", w, r)
	defer txn.End()
	ctx.render(r, w, "help.html", nil)
}

func (ctx *Context) HTTPHandler() http.Handler {
	router := vestigo.NewRouter()
	router.Get("/", ctx.HandleIndex)
	router.Post("/appraisal", ctx.HandleAppraisal)
	router.Post("/estimate", ctx.HandleAppraisal)
	router.Get("/a/:appraisalID", ctx.HandleViewAppraisal)
	router.Get("/e/:legacyAppraisalID", ctx.HandleViewAppraisal)
	router.Get("/item/:typeID", ctx.HandleViewItem)
	router.Get("/search", ctx.HandleSearch)
	router.Get("/latest", ctx.HandleLatestAppraisals)
	router.Get("/legal", ctx.HandleLegal)
	router.Get("/help", ctx.HandleHelp)

	vestigo.CustomNotFoundHandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx.renderErrorPage(r, w, http.StatusNotFound, "Not Found", "I couldn't find what you're looking for")
		})

	vestigo.CustomMethodNotAllowedHandlerFunc(func(allowedMethods string) func(w http.ResponseWriter, r *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx.renderErrorPage(r, w, http.StatusInternalServerError, "Method not allowed", fmt.Sprintf("HTTP Method not allowed. What is allowed is: "+allowedMethods))
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

	// Wrap global handlers
	handler := http.Handler(mux)
	handler = accesslog.NewLoggingHandler(handler, accessLogger{})
	handler = context.ClearHandler(handler)

	return handler
}
