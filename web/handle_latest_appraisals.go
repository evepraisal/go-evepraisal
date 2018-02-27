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

	opts := evepraisal.ListAppraisalsOptions{
		Limit:         int(limit),
		Kind:          r.FormValue("kind"),
		SortDirection: "DESC",
	}

	appraisals, err := ctx.App.AppraisalDB.ListAppraisals(opts)
	if err != nil {
		ctx.renderServerError(r, w, err)
		return
	}

	ctx.render(r, w, "latest.html", struct {
		Appraisals []evepraisal.Appraisal `json:"appraisals"`
	}{cleanAppraisals(appraisals)})
}
