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
	"github.com/pquerna/ffjson/ffjson"
)

var spewConfig = spew.ConfigState{Indent: "    ", SortKeys: true}
var templateFuncs = template.FuncMap{
	"humanizeVolume":  humanizeVolume,
	"comma":           humanize.Comma,
	"commaf":          humanizeCommaf,
	"commai":          func(i int) string { return humanize.Comma(int64(i)) },
	"prettybignumber": HumanLargeNumber,
	"relativetime":    humanize.Time,
	"timefmt":         func(t time.Time) string { return t.Format("2006-01-02 15:04:05") },

	// Math
	"divide":   func(a, b int64) float64 { return float64(a) / float64(b) },
	"multiply": func(a, b float64) float64 { return a * b },

	// Appraisal-specific
	"appraisallink":       appraisalLink,
	"normalAppraisalLink": normalAppraisalLink,
	"liveAppraisalLink":   liveAppraisalLink,
	"rawAppraisalLink":    rawAppraisalLink,
	"jsonAppraisalLink":   jsonAppraisalLink,

	// Only for debugging
	"spew": spewConfig.Sdump,
}

type namedThing struct {
	Name        string
	DisplayName string
}

var selectableMarkets = []namedThing{
	{Name: "jita", DisplayName: "Jita"},
	{Name: "universe", DisplayName: "Universe"},
	{Name: "amarr", DisplayName: "Amarr"},
	{Name: "dodixie", DisplayName: "Dodixie"},
	{Name: "hek", DisplayName: "Hek"},
	{Name: "rens", DisplayName: "Rens"},
}

var selectableVisibilities = []namedThing{
	{Name: "public", DisplayName: "Public"},
	{Name: "private", DisplayName: "Private"},
}

// PageRoot is basically the root of the page. It includes some details that are given on every page and page-specific data
type PageRoot struct {
	UI struct {
		SelectedMarket       string
		Markets              []namedThing
		SelectedVisibility   string
		Visibilities         []namedThing
		SelectedPersist      bool
		PricePercentage      float64
		ExpireAfter          string
		BaseURL              string
		BaseURLWithoutScheme string
		User                 *evepraisal.User
		LoginEnabled         bool
		RawTextAreaDefault   string
		FlashMessages        []FlashMessage
		Path                 string
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

	if r.Header.Get("format") == formatJSON {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err := ffjson.NewEncoder(w).Encode(root.Page)
		if err != nil {
			log.Println("Error when encoding JSON: ", err)
		}
	} else {
		root.UI.Path = r.URL.Path
		root.UI.SelectedMarket = ctx.getSessionValueWithDefault(r, "market", "jita")
		root.UI.Markets = selectableMarkets
		root.UI.SelectedVisibility = ctx.getSessionValueWithDefault(r, "visibility", "public")
		root.UI.Visibilities = selectableVisibilities
		root.UI.SelectedPersist = ctx.getSessionBooleanWithDefault(r, "persist", true)
		root.UI.PricePercentage = ctx.getSessionFloat64WithDefault(r, "price_percentage", 100)
		root.UI.ExpireAfter = ctx.getSessionValueWithDefault(r, "expire_after", "360h")
		root.UI.BaseURLWithoutScheme = strings.TrimPrefix(strings.TrimPrefix(ctx.BaseURL, "https://"), "http://")
		root.UI.BaseURL = ctx.BaseURL
		root.UI.FlashMessages = ctx.getFlashMessages(r, w)
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
	}

	return nil
}

func (ctx *Context) renderErrorPage(r *http.Request, w http.ResponseWriter, statusCode int, title, message string) {
	ctx.renderErrorPageWithRoot(r, w, statusCode, title, message, PageRoot{})
}

func (ctx *Context) renderErrorPageWithRoot(r *http.Request, w http.ResponseWriter, statusCode int, title, message string, root PageRoot) {
	root.Page = struct {
		ErrorTitle   string `json:"error_title"`
		ErrorMessage string `json:"error_message"`
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

// Reload will re-parse templates and set them on the context object
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

	root.New("extra-html-header").Parse(ctx.ExtraHTMLHeader)
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
