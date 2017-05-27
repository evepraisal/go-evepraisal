package web

import (
	"net/http"

	"github.com/evepraisal/go-evepraisal/typedb"
)

func (ctx *Context) HandleSearch(w http.ResponseWriter, r *http.Request) {
	txn := ctx.app.TransactionLogger.StartWebTransaction("search", w, r)
	defer txn.End()

	results := ctx.app.TypeDB.Search(r.FormValue("q"))
	ctx.render(r, w, "search.html", struct{ Results []typedb.EveType }{Results: results})
}
