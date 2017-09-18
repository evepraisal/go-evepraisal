package web

import (
	"net/http"
	"strconv"

	"github.com/evepraisal/go-evepraisal"
)

// HandleLatestAppraisals is the handler for /latest
func (ctx *Context) HandleLatestAppraisals(w http.ResponseWriter, r *http.Request) {
	var limit int64
	var err error
	limit, err = strconv.ParseInt(r.FormValue("limit"), 10, 64)
	if err != nil {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	appraisals, err := ctx.App.AppraisalDB.LatestAppraisals(int(limit), r.FormValue("kind"))
	if err != nil {
		ctx.renderServerError(r, w, err)
		return
	}

	ctx.render(r, w, "latest.html", struct {
		Appraisals []evepraisal.Appraisal `json:"appraisals"`
	}{appraisals})
}

// HandleUserLatestAppraisals is the handler for /user/latest
func (ctx *Context) HandleUserLatestAppraisals(w http.ResponseWriter, r *http.Request) {
	user := ctx.GetCurrentUser(r)
	if user == nil {
		ctx.renderErrorPage(r, w, http.StatusUnauthorized, "Not logged in", "You need to be logged in to see this page")
		return
	}

	var limit int64
	var err error
	limit, err = strconv.ParseInt(r.FormValue("limit"), 10, 64)
	if err != nil {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	appraisals, err := ctx.App.AppraisalDB.LatestAppraisalsByUser(*user, int(limit), r.FormValue("kind"))
	if err != nil {
		ctx.renderServerError(r, w, err)
		return
	}

	ctx.render(r, w, "user_latest.html", struct {
		Appraisals []evepraisal.Appraisal `json:"appraisals"`
	}{appraisals})
}
