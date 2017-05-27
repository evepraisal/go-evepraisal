package web

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/dustin/go-humanize"
)

var templateFuncs = template.FuncMap{
	"commaf":          humanizeCommaf,
	"comma":           humanize.Comma,
	"prettybignumber": HumanLargeNumber,
	"relativetime":    humanize.Time,
	"timefmt":         func(t time.Time) string { return t.Format("2006-01-02 15:04:05") },

	// Only for debugging
	"spew": spew.Sdump,
}

type displayMarket struct {
	Name        string
	DisplayName string
}

var selectableMarkets = []displayMarket{
	{Name: "jita", DisplayName: "Jita"},
	{Name: "universe", DisplayName: "Universe"},
	{Name: "amarr", DisplayName: "Amarr"},
	{Name: "dodixie", DisplayName: "Dodixie"},
	{Name: "rens", DisplayName: "Rens"},
	{Name: "hek", DisplayName: "Hek"},
}

type PageRoot struct {
	UI struct {
		SelectedMarket       string
		Markets              []displayMarket
		BaseURL              string
		BaseURLWithoutScheme string
	}
	Page interface{}
}

func (ctx *Context) render(r *http.Request, w http.ResponseWriter, templateName string, page interface{}) error {
	tmpl, ok := ctx.templates[templateName]
	if !ok {
		return fmt.Errorf("Could not find template named '%s'", templateName)
	}

	root := PageRoot{Page: page}
	root.UI.SelectedMarket = ctx.getDefaultMarket(r)
	root.UI.Markets = selectableMarkets

	root.UI.BaseURLWithoutScheme = strings.TrimPrefix(strings.TrimPrefix(ctx.baseURL, "https://"), "http://")
	root.UI.BaseURL = ctx.baseURL

	err := tmpl.ExecuteTemplate(w, templateName, root)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}

func (ctx *Context) renderErrorPage(r *http.Request, w http.ResponseWriter, statusCode int, title, message string) {
	w.WriteHeader(statusCode)
	ctx.render(r, w, "error.html", struct {
		ErrorTitle   string
		ErrorMessage string
	}{title, message})
}

func (ctx *Context) Reload() error {
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

	root.New("extra-javascript").Parse(ctx.extraJS)
	root.New("ad-block").Parse(ctx.adBlock)

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

	ctx.templates = templates
	return nil
}
