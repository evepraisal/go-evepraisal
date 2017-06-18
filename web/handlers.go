package web

import (
	"io"
	"log"
	"net/http"

	"github.com/NYTimes/gziphandler"
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/go-zoo/bone"
	"github.com/gorilla/context"
	"github.com/mash/go-accesslog"
	"github.com/newrelic/go-agent"
)

func (ctx *Context) HandleIndex(w http.ResponseWriter, r *http.Request) {
	total, err := ctx.App.AppraisalDB.TotalAppraisals()
	if err != nil {
		ctx.renderErrorPage(r, w, http.StatusInternalServerError, "Something bad happened", err.Error())
		return
	}
	ctx.render(r, w, "main.html", struct{ TotalAppraisalCount int64 }{TotalAppraisalCount: total})
}

func (ctx *Context) HandleLegal(w http.ResponseWriter, r *http.Request) {
	ctx.render(r, w, "legal.html", nil)
}

func (ctx *Context) HandleHelp(w http.ResponseWriter, r *http.Request) {
	ctx.render(r, w, "help.html", nil)
}

func (ctx *Context) HandleRobots(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain")
	io.WriteString(w, `User-agent: *
Disallow:`)
}

func (ctx *Context) HandleFavicon(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "static/favicon.ico", http.StatusPermanentRedirect)
}

func (ctx *Context) HTTPHandler() http.Handler {

	router := bone.New()
	router.GetFunc("/", ctx.HandleIndex)
	router.PostFunc("/appraisal", ctx.HandleAppraisal)
	router.PostFunc("/estimate", ctx.HandleAppraisal)
	router.GetFunc("/a/:appraisalID", ctx.HandleViewAppraisal)
	router.GetFunc("/e/:legacyAppraisalID", ctx.HandleViewAppraisal)
	router.GetFunc("/item/:typeID", ctx.HandleViewItem)
	router.GetFunc("/search", ctx.HandleSearch)
	router.GetFunc("/search.json", ctx.HandleSearchJSON)
	router.GetFunc("/latest", ctx.HandleLatestAppraisals)
	router.GetFunc("/legal", ctx.HandleLegal)
	router.GetFunc("/help", ctx.HandleHelp)
	router.GetFunc("/robots.txt", ctx.HandleRobots)
	router.GetFunc("/favicon.ico", ctx.HandleFavicon)

	// Authenticated pages
	router.GetFunc("/login", ctx.HandleLogin)
	router.GetFunc("/logout", ctx.HandleLogout)
	router.GetFunc("/oauthcallback", ctx.HandleAuthCallback)
	router.GetFunc("/user/latest", ctx.HandleUserLatestAppraisals)

	router.NotFound(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx.renderErrorPage(r, w, http.StatusNotFound, "Not Found", "I couldn't find what you're looking for")
	}))

	if ctx.App.NewRelicApplication != nil {
		for _, routes := range router.Routes {
			for _, route := range routes {
				_, h := newrelic.WrapHandle(ctx.App.NewRelicApplication, route.Path, route.Handler)
				route.Handler = h
			}
		}
	}

	mux := http.NewServeMux()
	setCacheHeaders := func(h http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Cache-Control", "public, max-age=3600")
			etag, ok := ctx.etags[r.RequestURI]
			if ok {
				w.Header().Add("Etag", etag)
			}
			h.ServeHTTP(w, r)
		}
	}

	// Route our bundled static files
	var fs = &assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo, Prefix: "/static/"}
	mux.Handle("/static/", setCacheHeaders(http.StripPrefix("/static/", http.FileServer(fs))))

	// Mount our web app router to root
	mux.Handle("/", router)
	err := ctx.Reload()
	if err != nil {
		log.Fatal(err)
	}

	// Wrap global handlers
	handler := http.Handler(mux)
	alogger := accessLogger{}
	userLoggerInjectHandler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := ctx.GetCurrentUser(r)
			if user != nil {
				r.Header.Set("logged-in-user", user.CharacterName)
			} else {
				r.Header.Set("logged-in-user", "")
			}

			next.ServeHTTP(w, r)
		})
	}

	handler = userLoggerInjectHandler(handler)
	handler = accesslog.NewLoggingHandler(handler, alogger)
	handler = context.ClearHandler(handler)
	handler = gziphandler.GzipHandler(handler)

	return handler
}
