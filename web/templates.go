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
	"github.com/evepraisal/go-evepraisal"
)

var templateFuncs = template.FuncMap{
	"humanizeVolume":  humanizeVolume,
	"commaf":          humanizeCommaf,
	"comma":           humanize.Comma,
	"prettybignumber": HumanLargeNumber,
	"relativetime":    humanize.Time,
	"timefmt":         func(t time.Time) string { return t.Format("2006-01-02 15:04:05") },

	// Math
	"divide":   func(a, b int64) float64 { return float64(a) / float64(b) },
	"multiply": func(a, b float64) float64 { return a * b },

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
	{Name: "hek", DisplayName: "Hek"},
}

type PageRoot struct {
	UI struct {
		SelectedMarket       string
		Markets              []displayMarket
		BaseURL              string
		BaseURLWithoutScheme string
		User                 *evepraisal.User
		LoginEnabled         bool
		RawTextAreaDefault   string
	}
	Page interface{}
}

func (ctx *Context) render(r *http.Request, w http.ResponseWriter, templateName string, page interface{}) error {
	return ctx.renderWithRoot(r, w, templateName, PageRoot{Page: page})
}

func (ctx *Context) renderWithRoot(r *http.Request, w http.ResponseWriter, templateName string, root PageRoot) error {
	tmpl, ok := ctx.templates[templateName]
	if !ok {
		return fmt.Errorf("Could not find template named '%s'", templateName)
	}

	root.UI.SelectedMarket = ctx.getDefaultMarket(r)
	root.UI.Markets = selectableMarkets
	root.UI.BaseURLWithoutScheme = strings.TrimPrefix(strings.TrimPrefix(ctx.BaseURL, "https://"), "http://")
	root.UI.BaseURL = ctx.BaseURL
	if ctx.OauthConfig != nil {
		root.UI.LoginEnabled = true
		root.UI.User = ctx.GetCurrentUser(r)
	}

	w.Header().Add("Content-Type", "text/html")
	err := tmpl.ExecuteTemplate(w, templateName, root)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}

func (ctx *Context) renderErrorPage(r *http.Request, w http.ResponseWriter, statusCode int, title, message string) {
	ctx.renderErrorPageWithRoot(r, w, statusCode, title, message, PageRoot{})
}

func (ctx *Context) renderErrorPageWithRoot(r *http.Request, w http.ResponseWriter, statusCode int, title, message string, root PageRoot) {
	root.Page = struct {
		ErrorTitle   string
		ErrorMessage string
	}{title, message}
	w.WriteHeader(statusCode)
	ctx.renderWithRoot(r, w, "error.html", root)
}

func (ctx *Context) renderServerError(r *http.Request, w http.ResponseWriter, err error) {
	ctx.renderServerErrorWithRoot(r, w, err, PageRoot{})
}

func (ctx *Context) renderServerErrorWithRoot(r *http.Request, w http.ResponseWriter, err error, root PageRoot) {
	log.Printf("ERROR: %s", err)
	ctx.renderErrorPageWithRoot(r, w, http.StatusInternalServerError, "Something bad happened", err.Error(), root)
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

	root.New("extra-javascript").Parse(ctx.ExtraJS)
	root.New("ad-block").Parse(ctx.AdBlock)

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
