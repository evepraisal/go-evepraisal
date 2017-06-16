package web

import (
	"html/template"

	"golang.org/x/oauth2"

	"github.com/evepraisal/go-evepraisal"
	"github.com/gorilla/sessions"
)

type Context struct {
	App            *evepraisal.App
	BaseURL        string
	ExtraJS        string
	AdBlock        string
	CookieStore    *sessions.CookieStore
	OauthConfig    *oauth2.Config
	OauthVerifyURL string

	templates map[string]*template.Template
	etags     map[string]string
}

func NewContext(app *evepraisal.App) *Context {
	ctx := &Context{App: app}
	ctx.GenerateStaticEtags()
	return ctx
}
