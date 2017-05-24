package web

import (
	"html/template"

	"github.com/evepraisal/go-evepraisal"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

type Context struct {
	app         *evepraisal.App
	baseURL     string
	extraJS     string
	templates   map[string]*template.Template
	cookieStore *sessions.CookieStore
}

func NewContext(app *evepraisal.App, baseURL string, extraJS string) *Context {
	return &Context{
		app:         app,
		baseURL:     baseURL,
		extraJS:     extraJS,
		cookieStore: sessions.NewCookieStore(securecookie.GenerateRandomKey(32)),
	}
}
