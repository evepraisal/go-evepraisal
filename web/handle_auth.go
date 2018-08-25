package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/evepraisal/go-evepraisal"

	"golang.org/x/oauth2"
)

// GetCurrentUser returns the current user, if there is one in the request cookies
func (ctx *Context) GetCurrentUser(r *http.Request) *evepraisal.User {
	user, ok := ctx.getSessionValue(r, "user").(evepraisal.User)
	if !ok {
		return nil
	}
	return &user
}

// IsAppraisalOwner determines if a user owns the given appraisal
func IsAppraisalOwner(user *evepraisal.User, appraisal *evepraisal.Appraisal) bool {
	if user == nil {
		return false
	}

	if appraisal.User == nil {
		return false
	}

	return appraisal.User.CharacterOwnerHash == user.CharacterOwnerHash
}

// HandleLogin handles /login
func (ctx *Context) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if ctx.OauthConfig == nil {
		ctx.renderErrorPage(r, w, http.StatusBadRequest, "Feature Unavailable", "SSO is not configured")
		return
	}
	url := ctx.OauthConfig.AuthCodeURL("state", oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// HandleLogout handles /logout
func (ctx *Context) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if ctx.OauthConfig == nil {
		ctx.renderErrorPage(r, w, http.StatusBadRequest, "Feature Unavailable", "SSO is not configured")
		return
	}
	ctx.setSessionValue(r, w, "user", nil)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// HandleAuthCallback handles /oauthcallback
func (ctx *Context) HandleAuthCallback(w http.ResponseWriter, r *http.Request) {
	if ctx.OauthConfig == nil {
		ctx.renderErrorPage(r, w, http.StatusBadRequest, "Feature Unavailable", "SSO is not configured")
		return
	}
	code := r.FormValue("code")
	tok, err := ctx.OauthConfig.Exchange(r.Context(), code)
	if err != nil {
		ctx.renderServerError(r, w, err)
	}

	client := ctx.OauthConfig.Client(r.Context(), tok)
	resp, err := client.Get(ctx.OauthVerifyURL)
	if err != nil {
		ctx.renderServerError(r, w, err)
		return
	}

	if resp.StatusCode != 200 {
		ctx.renderErrorPage(
			r, w, http.StatusInternalServerError,
			"Something bad happened",
			fmt.Sprintf("Invalid response when verifying identity (%d)", resp.StatusCode))
		return
	}

	user := &evepraisal.User{}
	err = json.NewDecoder(resp.Body).Decode(user)
	defer resp.Body.Close()
	if err != nil {
		ctx.renderServerError(r, w, err)
		return
	}

	ctx.setSessionValue(r, w, "user", &user)
	log.Printf("User logged in: %s", user.CharacterName)

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
