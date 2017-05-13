package evepraisal

//go:generate $GOPATH/bin/go-bindata --pkg evepraisal -prefix resources/ resources/...

import (
	"encoding/json"
	"expvar"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/dustin/go-humanize"
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/husobee/vestigo"
	"github.com/mash/go-accesslog"
)

type accessLogger struct {
}

func (l accessLogger) Log(record accesslog.LogRecord) {
	log.Printf("%s %s %d (%s) - %d", record.Method, record.Uri, record.Status, record.Ip, record.Size)
}

var templateFuncs = template.FuncMap{
	"commaf":          humanize.Commaf,
	"comma":           humanize.Comma,
	"spew":            spew.Sdump,
	"prettybignumber": HumanLargeNumber,
	"relativetime":    humanize.Time,
	"timefmt":         func(t time.Time) string { return t.Format("2006-01-02 15:04:05") },
}
var templates = MustLoadTemplateFiles()

func MustLoadTemplateFiles() *template.Template {
	t := template.New("root").Funcs(templateFuncs)
	for _, path := range AssetNames() {
		if strings.HasPrefix(path, "templates/") {
			tmpl := t.New(strings.TrimPrefix(path, "templates/"))
			fileContents, err := Asset(path)
			if err != nil {
				panic(err)
			}

			_, err = tmpl.Parse(string(fileContents))
			if err != nil {
				panic(err)
			}
		}
	}
	return t
}

type MainPageStruct struct {
	Appraisal *Appraisal
}

func (app *App) HandleIndex(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "main.html", MainPageStruct{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *App) HandleAppraisal(w http.ResponseWriter, r *http.Request) {
	appraisal, err := app.StringToAppraisal(r.FormValue("body"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		templates.ExecuteTemplate(w, "error.html", ErrorPage{
			ErrorTitle:   "Invalid input",
			ErrorMessage: err.Error(),
		})
		return
	}

	err = app.AppraisalDB.PutNewAppraisal(appraisal)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		templates.ExecuteTemplate(w, "error.html", ErrorPage{
			ErrorTitle:   "Error when storing appraisal",
			ErrorMessage: err.Error(),
		})
		return
	}

	err = templates.ExecuteTemplate(
		w,
		"main.html",
		MainPageStruct{Appraisal: appraisal})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *App) HandleViewAppraisal(w http.ResponseWriter, r *http.Request) {

	appraisalID := vestigo.Param(r, "appraisalID")
	if strings.HasSuffix(appraisalID, ".json") {
		app.HandleViewAppraisalJSON(w, r)
		return
	}

	appraisal, err := app.AppraisalDB.GetAppraisal(appraisalID)
	if err == AppraisalNotFound {
		w.WriteHeader(http.StatusNotFound)
		templates.ExecuteTemplate(w, "error.html", ErrorPage{
			ErrorTitle:   "Not Found",
			ErrorMessage: "I couldn't find what you're looking for",
		})
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		templates.ExecuteTemplate(w, "error.html", ErrorPage{
			ErrorTitle:   "Something bad happened",
			ErrorMessage: err.Error(),
		})
		return
	}

	err = templates.ExecuteTemplate(
		w,
		"main.html",
		MainPageStruct{Appraisal: appraisal})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *App) HandleViewAppraisalJSON(w http.ResponseWriter, r *http.Request) {
	appraisalID := vestigo.Param(r, "appraisalID")
	appraisalID = strings.TrimSuffix(appraisalID, ".json")

	appraisal, err := app.AppraisalDB.GetAppraisal(appraisalID)
	if err == AppraisalNotFound {
		w.WriteHeader(http.StatusNotFound)
		templates.ExecuteTemplate(w, "error.html", ErrorPage{
			ErrorTitle:   "Not Found",
			ErrorMessage: "I couldn't find what you're looking for",
		})
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		templates.ExecuteTemplate(w, "error.html", ErrorPage{
			ErrorTitle:   "Something bad happened",
			ErrorMessage: err.Error(),
		})
		return
	}

	r.Header["Content-Type"] = []string{"application/json"}
	json.NewEncoder(w).Encode(appraisal)
}

func (app *App) HandleLatestAppraisals(w http.ResponseWriter, r *http.Request) {
	appraisals, err := app.AppraisalDB.LatestAppraisals(100, "")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		templates.ExecuteTemplate(w, "error.html", ErrorPage{
			ErrorTitle:   "Something bad happened",
			ErrorMessage: err.Error(),
		})
		return
	}

	err = templates.ExecuteTemplate(
		w,
		"latest.html",
		struct{ Appraisals []Appraisal }{Appraisals: appraisals})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type ErrorPage struct {
	ErrorTitle   string
	ErrorMessage string
}

func HTTPHandler(app *App) http.Handler {
	router := vestigo.NewRouter()
	router.Get("/latest", app.HandleLatestAppraisals)
	router.Get("/", app.HandleIndex)
	router.Post("/", app.HandleAppraisal)
	router.Get("/a/:appraisalID", app.HandleViewAppraisal)

	router.Handle("/expvar", expvar.Handler())

	vestigo.CustomNotFoundHandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			templates.ExecuteTemplate(w, "error.html", ErrorPage{
				ErrorTitle:   "Not Found",
				ErrorMessage: "I couldn't find what you're looking for",
			})
		})

	vestigo.CustomMethodNotAllowedHandlerFunc(func(allowedMethods string) func(w http.ResponseWriter, r *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusMethodNotAllowed)
			templates.ExecuteTemplate(w, "error.html", ErrorPage{
				ErrorTitle:   "Method not allowed",
				ErrorMessage: fmt.Sprintf("HTTP Method not allowed. What is allowed is: " + allowedMethods),
			})
		}
	})

	mux := http.NewServeMux()

	// Route our bundled static files
	var fs = &assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo, Prefix: "/static/"}
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(fs)))

	// Mount our web app router to root
	mux.Handle("/", router)

	// Setup access logger
	l := accessLogger{}

	return accesslog.NewLoggingHandler(mux, l)
}
