package web

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/elazarl/go-bindata-assetfs"
	"github.com/evepraisal/go-evepraisal"
	"github.com/husobee/vestigo"
	"github.com/mash/go-accesslog"
)

type Context struct {
	app       *evepraisal.App
	extraJS   string
	templates map[string]*template.Template
}

func NewContext(app *evepraisal.App, extraJS string) *Context {
	return &Context{
		app:     app,
		extraJS: extraJS,
	}
}

func (ctx *Context) HTTPHandler() http.Handler {
	router := vestigo.NewRouter()
	router.Get("/", ctx.HandleIndex)
	router.Post("/", ctx.HandleAppraisal)
	router.Get("/a/:appraisalID", ctx.HandleViewAppraisal)
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
