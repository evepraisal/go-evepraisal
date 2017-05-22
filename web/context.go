package web

import (
	"html/template"

	"github.com/evepraisal/go-evepraisal"
)

type Context struct {
	app       *evepraisal.App
	extraJS   string
	templates map[string]*template.Template
}

func NewContext(app *evepraisal.App, extraJS string) *Context {
	return &Context{
		app:     app,
		extraJS: extraJS,
	}
}
