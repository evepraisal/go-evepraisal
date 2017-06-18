package web

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/NYTimes/gziphandler"
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/context"
	"github.com/husobee/vestigo"
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

	router := vestigo.NewRouter()
	addHandler := func(method string, pattern string, handler http.HandlerFunc) {
		if ctx.App.NewRelicApplication != nil {
			_, h := newrelic.WrapHandle(ctx.App.NewRelicApplication, pattern, http.HandlerFunc(handler))
			handler = func(w http.ResponseWriter, r *http.Request) { h.ServeHTTP(w, r) }
		}
		router.Add(method, pattern, handler)
	}

	addHandler("GET", "/", ctx.HandleIndex)
	addHandler("POST", "/appraisal", ctx.HandleAppraisal)
	addHandler("POST", "/estimate", ctx.HandleAppraisal)
	addHandler("GET", "/a/:appraisalID", ctx.HandleViewAppraisal)
	addHandler("GET", "/e/:legacyAppraisalID", ctx.HandleViewAppraisal)
	addHandler("GET", "/item/:typeID", ctx.HandleViewItem)
	addHandler("GET", "/search", ctx.HandleSearch)
	addHandler("GET", "/search.json", ctx.HandleSearchJSON)
	addHandler("GET", "/latest", ctx.HandleLatestAppraisals)
	addHandler("GET", "/legal", ctx.HandleLegal)
	addHandler("GET", "/help", ctx.HandleHelp)
	addHandler("GET", "/robots.txt", ctx.HandleRobots)
	addHandler("GET", "/favicon.ico", ctx.HandleFavicon)

	// Authenticated pages
	addHandler("GET", "/login", ctx.HandleLogin)
	addHandler("GET", "/logout", ctx.HandleLogout)
	addHandler("GET", "/oauthcallback", ctx.HandleAuthCallback)
	addHandler("GET", "/user/latest", ctx.HandleUserLatestAppraisals)

	vestigo.CustomNotFoundHandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx.renderErrorPage(r, w, http.StatusNotFound, "Not Found", "I couldn't find what you're looking for")
		})

	vestigo.CustomMethodNotAllowedHandlerFunc(func(allowedMethods string) func(w http.ResponseWriter, r *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx.renderErrorPage(r, w, http.StatusMethodNotAllowed, "Method not allowed", fmt.Sprintf("HTTP Method not allowed. What is allowed is: "+allowedMethods))
		}
	})

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
