package evepraisal

//go:generate go-bindata --pkg evepraisal -prefix resources/ resources/...

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/elazarl/go-bindata-assetfs"
	"github.com/husobee/vestigo"
)

var serverPort = 8080
var templates = MustLoadTemplateFiles()

func MustLoadTemplateFiles() *template.Template {
	t := template.New("root")
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

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "main.html", struct{}{})
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
	json.NewEncoder(w).Encode(appraisal)
}

func HTTPServer(addr string) *http.Server {
	log.Println("Included assets:")
	assets := AssetNames()
	sort.Strings(assets)
	for _, filename := range assets {
		log.Println(" - ", filename)
	}

	router := vestigo.NewRouter()
	router.Get("/", IndexHandler)
	router.Post("/appraise", AppraiseHandler)

	mux := http.NewServeMux()

	// Route our bundled static files
	var fs = &assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo, Prefix: "/static/"}
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(fs)))

	// Mount our web app router to root
	mux.Handle("/", router)

	return &http.Server{Addr: addr, Handler: mux}
}
