package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/evepraisal/go-evepraisal"

	"golang.org/x/oauth2"
)

func (ctx *Context) GetCurrentUser(r *http.Request) *evepraisal.User {
	user, ok := ctx.getSessionValue(r, "user").(evepraisal.User)
	if !ok {
		return nil
	}
	return &user
}

func (ctx *Context) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if ctx.OauthConfig == nil {
		ctx.renderErrorPage(r, w, http.StatusBadRequest, "Feature Unavailable", "SSO is not configured")
		return
	}
	url := ctx.OauthConfig.AuthCodeURL("state", oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (ctx *Context) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if ctx.OauthConfig == nil {
		ctx.renderErrorPage(r, w, http.StatusBadRequest, "Feature Unavailable", "SSO is not configured")
		return
	}
	ctx.setSessionValue(r, w, "user", nil)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (ctx *Context) HandleAuthCallback(w http.ResponseWriter, r *http.Request) {
	if ctx.OauthConfig == nil {
		ctx.renderErrorPage(r, w, http.StatusBadRequest, "Feature Unavailable", "SSO is not configured")
		return
	}
	code := r.FormValue("code")
	tok, err := ctx.OauthConfig.Exchange(r.Context(), code)
	if err != nil {
		ctx.renderErrorPage(r, w, http.StatusInternalServerError, "Something bad happened", err.Error())
	}

	client := ctx.OauthConfig.Client(r.Context(), tok)
	resp, err := client.Get(ctx.OauthVerifyURL)
	if err != nil {
		ctx.renderErrorPage(r, w, http.StatusInternalServerError, "Something bad happened", err.Error())
	}

	if resp.StatusCode != 200 {
		ctx.renderErrorPage(r, w, http.StatusInternalServerError, "Something bad happened", fmt.Sprint("Invalid response when verifying identity (%d)", resp.StatusCode))
	}

	user := &evepraisal.User{}
	err = json.NewDecoder(resp.Body).Decode(user)
	defer resp.Body.Close()
	if err != nil {
		ctx.renderErrorPage(r, w, http.StatusInternalServerError, "Something bad happened", err.Error())
	}

	ctx.setSessionValue(r, w, "user", &user)
	log.Printf("User logged in: %s", user.CharacterName)

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
