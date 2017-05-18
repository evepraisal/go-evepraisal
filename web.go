package evepraisal

import (
	"encoding/json"
	"expvar"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"sort"
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
	"commaf":          humanizeCommaf,
	"comma":           humanize.Comma,
	"prettybignumber": HumanLargeNumber,
	"relativetime":    humanize.Time,
	"timefmt":         func(t time.Time) string { return t.Format("2006-01-02 15:04:05") },

	// Only for debugging
	"spew": spew.Sdump,
}

func (app *App) LoadTemplates() error {

	templates := make(map[string]*template.Template)
	root := template.New("root").Funcs(templateFuncs)

	for _, path := range AssetNames() {
		if strings.HasPrefix(path, "templates/") && strings.HasPrefix(filepath.Base(path), "_") {
			log.Println("load partial:", path)
			tmplPartial := root.New(strings.TrimPrefix(path, "templates/"))
			fileContents, err := Asset(path)
			if err != nil {
				return err
			}
			_, err = tmplPartial.Parse(string(fileContents))
			if err != nil {
				return err
			}
		}
	}

	root.New("extra-javascript").Parse(app.ExtraJS)

	for _, path := range AssetNames() {
		baseName := filepath.Base(path)
		if strings.HasPrefix(path, "templates/") && !strings.HasPrefix(baseName, "_") {
			log.Println("load:", baseName)
			r, err := root.Clone()
			if err != nil {
				return err
			}
			tmpl := r.New(strings.TrimPrefix(path, "templates/"))
			fileContents, err := Asset(path)
			if err != nil {
				return err
			}

			_, err = tmpl.Parse(string(fileContents))
			if err != nil {
				return err
			}
			templates[baseName] = tmpl
		}
	}

	for _, template := range root.Templates() {
		log.Println(template.Name())
	}

	app.templates = templates
	return nil
}

func (app *App) render(w http.ResponseWriter, templateName string, input interface{}) error {
	tmpl, ok := app.templates[templateName]
	if !ok {
		return fmt.Errorf("Could not find template named '%s'", templateName)
	}
	err := tmpl.ExecuteTemplate(w, templateName, input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}

type MainPageStruct struct {
	Appraisal           *Appraisal
	TotalAppraisalCount int64
}

func (app *App) HandleIndex(w http.ResponseWriter, r *http.Request) {
	txn := app.TransactionLogger.StartWebTransaction("view_index", w, r)
	defer txn.End()

	total, err := app.AppraisalDB.TotalAppraisals()
	if err != nil {
		app.renderErrorPage(w, http.StatusInternalServerError, "Something bad happened", err.Error())
		return
	}
	err = app.render(w, "main.html", MainPageStruct{
		TotalAppraisalCount: total,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *App) HandleAppraisal(w http.ResponseWriter, r *http.Request) {
	txn := app.TransactionLogger.StartWebTransaction("create_appraisal", w, r)
	defer txn.End()

	log.Println("New appraisal at ", r.FormValue("market"))
	appraisal, err := app.StringToAppraisal(r.FormValue("market"), r.FormValue("body"))
	if err != nil {
		app.renderErrorPage(w, http.StatusBadRequest, "Invalid input", err.Error())
		return
	}

	err = app.AppraisalDB.PutNewAppraisal(appraisal)
	if err != nil {
		app.renderErrorPage(w, http.StatusInternalServerError, "Something bad happened", err.Error())
		return
	}

	err = app.render(
		w,
		"main.html",
		MainPageStruct{Appraisal: appraisal})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *App) HandleViewAppraisal(w http.ResponseWriter, r *http.Request) {
	txn := app.TransactionLogger.StartWebTransaction("view_appraisal", w, r)
	defer txn.End()

	appraisalID := vestigo.Param(r, "appraisalID")
	if strings.HasSuffix(appraisalID, ".json") {
		app.HandleViewAppraisalJSON(w, r)
		return
	}

	if strings.HasSuffix(appraisalID, ".raw") {
		app.HandleViewAppraisalRAW(w, r)
		return
	}

	appraisal, err := app.AppraisalDB.GetAppraisal(appraisalID)
	if err == AppraisalNotFound {
		app.renderErrorPage(w, http.StatusNotFound, "Not Found", "I couldn't find what you're looking for")
		return
	} else if err != nil {
		app.renderErrorPage(w, http.StatusInternalServerError, "Something bad happened", err.Error())
		return
	}

	sort.Slice(appraisal.Items, func(i, j int) bool {
		return appraisal.Items[i].SingleRepresentativePrice() > appraisal.Items[j].SingleRepresentativePrice()
	})

	err = app.render(
		w,
		"main.html",
		MainPageStruct{Appraisal: appraisal})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *App) HandleViewAppraisalJSON(w http.ResponseWriter, r *http.Request) {
	txn := app.TransactionLogger.StartWebTransaction("view_appraisal_json", w, r)
	defer txn.End()

	appraisalID := vestigo.Param(r, "appraisalID")
	appraisalID = strings.TrimSuffix(appraisalID, ".json")

	appraisal, err := app.AppraisalDB.GetAppraisal(appraisalID)
	if err == AppraisalNotFound {
		app.renderErrorPage(w, http.StatusNotFound, "Not Found", "I couldn't find what you're looking for")
		return
	} else if err != nil {
		app.renderErrorPage(w, http.StatusInternalServerError, "Something bad happened", err.Error())
		return
	}

	r.Header["Content-Type"] = []string{"application/json"}
	json.NewEncoder(w).Encode(appraisal)
}

func (app *App) HandleViewAppraisalRAW(w http.ResponseWriter, r *http.Request) {
	txn := app.TransactionLogger.StartWebTransaction("view_appraisal_raw", w, r)
	defer txn.End()

	appraisalID := vestigo.Param(r, "appraisalID")
	appraisalID = strings.TrimSuffix(appraisalID, ".raw")

	appraisal, err := app.AppraisalDB.GetAppraisal(appraisalID)
	if err == AppraisalNotFound {
		app.renderErrorPage(w, http.StatusNotFound, "Not Found", "I couldn't find what you're looking for")
		return
	} else if err != nil {
		app.renderErrorPage(w, http.StatusInternalServerError, "Something bad happened", err.Error())
		return
	}

	r.Header["Content-Type"] = []string{"application/text"}
	io.WriteString(w, appraisal.Raw)
}

func (app *App) HandleLatestAppraisals(w http.ResponseWriter, r *http.Request) {
	txn := app.TransactionLogger.StartWebTransaction("view_latest_appraisals", w, r)
	defer txn.End()

	appraisals, err := app.AppraisalDB.LatestAppraisals(100, "")
	if err != nil {
		app.renderErrorPage(w, http.StatusInternalServerError, "Something bad happened", err.Error())
		return
	}
	// {{ define "title"}}<title>Index Page</title>{{ end }}

	err = app.render(
		w,
		"latest.html",
		struct{ Appraisals []Appraisal }{appraisals})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *App) renderErrorPage(w http.ResponseWriter, statusCode int, title, message string) {
	w.WriteHeader(statusCode)
	app.render(w, "error.html", struct {
		ErrorTitle   string
		ErrorMessage string
	}{title, message})
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
			app.renderErrorPage(w, http.StatusNotFound, "Not Found", "I couldn't find what you're looking for")
		})

	vestigo.CustomMethodNotAllowedHandlerFunc(func(allowedMethods string) func(w http.ResponseWriter, r *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			app.renderErrorPage(w, http.StatusInternalServerError, "Method not allowed", fmt.Sprintf("HTTP Method not allowed. What is allowed is: "+allowedMethods))
		}
	})

	mux := http.NewServeMux()

	// Route our bundled static files
	var fs = &assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo, Prefix: "/static/"}
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(fs)))

	// Mount our web app router to root
	mux.Handle("/", router)
	err := app.LoadTemplates()
	if err != nil {
		log.Fatal(err)
	}

	return accesslog.NewLoggingHandler(mux, accessLogger{})
}
