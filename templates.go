package evepraisal

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

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

func (app *App) renderErrorPage(w http.ResponseWriter, statusCode int, title, message string) {
	w.WriteHeader(statusCode)
	app.render(w, "error.html", struct {
		ErrorTitle   string
		ErrorMessage string
	}{title, message})
}
