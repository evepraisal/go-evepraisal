package web

import (
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/NYTimes/gziphandler"
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/go-zoo/bone"
	"github.com/gorilla/context"
	"github.com/mash/go-accesslog"
	"github.com/newrelic/go-agent"
)

// HandleIndex is the handler for /
func (ctx *Context) HandleIndex(w http.ResponseWriter, r *http.Request) {
	total, err := ctx.App.AppraisalDB.TotalAppraisals()
	if err != nil {
		ctx.renderServerError(r, w, err)
		return
	}
	ctx.render(r, w, "main.html", struct {
		TotalAppraisalCount int64 `json:"total_appraisal_count"`
	}{TotalAppraisalCount: total})
}

// HandleLegal is the handler for /legal
func (ctx *Context) HandleLegal(w http.ResponseWriter, r *http.Request) {
	ctx.render(r, w, "legal.html", nil)
}

// HandleAbout is the handler for /about
func (ctx *Context) HandleAbout(w http.ResponseWriter, r *http.Request) {
	ctx.render(r, w, "about.html", nil)
}

// HandleAboutAPI is the handler for /about/api
func (ctx *Context) HandleAboutAPI(w http.ResponseWriter, r *http.Request) {
	ctx.render(r, w, "api.html", nil)
}

// HandleRobots is the handler for /robots.txt
func (ctx *Context) HandleRobots(w http.ResponseWriter, r *http.Request) {
	r.Header["Content-Type"] = []string{"text/plain"}
	io.WriteString(w, `User-agent: *
Disallow:`)
}

// HandleFavicon is the handler for /favicon.ico
func (ctx *Context) HandleFavicon(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "static/favicon.ico", http.StatusPermanentRedirect)
}

// HTTPHandler returns all HTTP handlers for the app
func (ctx *Context) HTTPHandler() http.Handler {

	router := bone.New()
	router.GetFunc("/", ctx.HandleIndex)

	// Create Appraisal
	router.PostFunc("/appraisal", ctx.HandleAppraisal)
	router.PostFunc("/estimate", ctx.HandleAppraisal)

	// Lates Appraisals
	router.GetFunc("/latest", ctx.HandleLatestAppraisals)

	// View Appraisal
	router.GetFunc("/a/#appraisalID^[a-zA-Z0-9]+$", ctx.HandleViewAppraisal)
	router.GetFunc("/a/#appraisalID^[a-zA-Z0-9]+$/#privateToken^[a-zA-Z0-9]+$", ctx.HandleViewAppraisal)
	router.GetFunc("/e/#legacyAppraisalID^[0-9]+$", ctx.HandleViewAppraisal)

	// View Item
	router.GetFunc("/item/#typeID^[0-9]$", ctx.HandleViewItem)

	// Search
	router.GetFunc("/search", ctx.HandleSearch)

	// Misc
	router.GetFunc("/legal", ctx.HandleLegal)
	router.GetFunc("/about", ctx.HandleAbout)
	router.GetFunc("/about/api", ctx.HandleAboutAPI)
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

	formatHandler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, ".json") {
				r.URL.Path = strings.TrimSuffix(r.URL.Path, ".json")
				r.Header.Set("format", "json")
			} else if strings.HasSuffix(r.URL.Path, ".raw") {
				r.URL.Path = strings.TrimSuffix(r.URL.Path, ".raw")
				r.Header.Set("format", "raw")
			} else {
				r.Header.Set("format", "")
			}
			next.ServeHTTP(w, r)
		})
	}

	handler = userLoggerInjectHandler(handler)
	handler = formatHandler(handler)
	handler = accesslog.NewLoggingHandler(handler, alogger)
	handler = context.ClearHandler(handler)
	handler = gziphandler.GzipHandler(handler)

	return handler
}
