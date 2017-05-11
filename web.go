package evepraisal

//go:generate $GOPATH/bin/go-bindata --pkg evepraisal -prefix resources/ resources/...

import (
	"expvar"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/dustin/go-humanize"
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/husobee/vestigo"
	"github.com/mash/go-accesslog"
	"github.com/spf13/viper"
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

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "main.html", MainPageStruct{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func AppraiseHandler(w http.ResponseWriter, r *http.Request) {
	appraisal, err := StringToAppraisal(r.FormValue("body"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

func HTTPServer() *http.Server {
	log.Println("Included assets:")
	assets := AssetNames()
	sort.Strings(assets)

	for _, filename := range assets {
		log.Printf(" -  %s", filename)
	}

	router := vestigo.NewRouter()
	router.Get("/", IndexHandler)
	router.Post("/appraise", AppraiseHandler)
	router.Handle("/expvar", expvar.Handler())

	mux := http.NewServeMux()

	// Route our bundled static files
	var fs = &assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo, Prefix: "/static/"}
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(fs)))

	// Mount our web app router to root
	mux.Handle("/", router)

	// Setup access logger
	l := accessLogger{}

	return &http.Server{Addr: viper.GetString("web.addr"), Handler: accesslog.NewLoggingHandler(mux, l)}
}
