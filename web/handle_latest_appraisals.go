package web

import (
	"net/http"
	"strconv"

	"github.com/evepraisal/go-evepraisal"
)

func (ctx *Context) HandleLatestAppraisals(w http.ResponseWriter, r *http.Request) {
	txn := ctx.app.TransactionLogger.StartWebTransaction("view_latest_appraisals", w, r)
	defer txn.End()

	var limit int64
	var err error
	limit, err = strconv.ParseInt(r.FormValue("limit"), 10, 64)
	if err != nil {
		limit = 100
	}

	appraisals, err := ctx.app.AppraisalDB.LatestAppraisals(int(limit), r.FormValue("kind"))
	if err != nil {
		ctx.renderErrorPage(r, w, http.StatusInternalServerError, "Something bad happened", err.Error())
		return
	}

	ctx.render(r, w, "latest.html", struct{ Appraisals []evepraisal.Appraisal }{appraisals})
}
