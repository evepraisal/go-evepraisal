package web

import (
	"fmt"
	"net/http"

	"github.com/evepraisal/go-evepraisal"
)

// HandleUserHistoryAppraisals is the handler for /user/latest
func (ctx *Context) HandleUserHistoryAppraisals(w http.ResponseWriter, r *http.Request) {
	user := ctx.GetCurrentUser(r)
	before := r.FormValue("before")
	limit := 50

	opts := evepraisal.ListAppraisalsOptions{
		Limit:          limit + 1,
		Kind:           r.FormValue("kind"),
		User:           user,
		EndAppraisalID: before,
		SortDirection:  "DESC",
	}

	appraisals, err := ctx.App.AppraisalDB.ListAppraisals(opts)
	if err != nil {
		ctx.renderServerError(r, w, err)
		return
	}

	hasMore := len(appraisals) > limit
	var next string
	if hasMore {
		next = fmt.Sprintf("/user/history?before=%s", appraisals[len(appraisals)-2].ID)
	}

	cleanAppraisals := cleanAppraisals(appraisals)
	// Cut off the extra one appraisal that we may have gotten (to see if there's another page)
	if len(appraisals) > limit {
		cleanAppraisals = cleanAppraisals[0:limit]
	}

	ctx.render(r, w, "user_history.html", struct {
		Appraisals []evepraisal.Appraisal `json:"appraisals"`
		Before     string                 `json:"before"`
		Limit      int                    `json:"limit"`
		HasMore    bool                   `json:"has_more"`
		Next       string                 `json:"next"`
	}{
		cleanAppraisals,
		before,
		limit,
		hasMore,
		next,
	})
}
