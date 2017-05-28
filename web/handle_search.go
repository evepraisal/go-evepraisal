package web

import (
	"encoding/json"
	"net/http"

	"github.com/evepraisal/go-evepraisal/typedb"
)

type SearchPage struct {
	Results []typedb.EveType `json:"results"`
}

func (ctx *Context) HandleSearch(w http.ResponseWriter, r *http.Request) {
	txn := ctx.app.TransactionLogger.StartWebTransaction("search", w, r)
	defer txn.End()

	results := ctx.app.TypeDB.Search(r.FormValue("q"))
	ctx.render(r, w, "search.html", SearchPage{Results: results})
}

func (ctx *Context) HandleSearchJSON(w http.ResponseWriter, r *http.Request) {
	txn := ctx.app.TransactionLogger.StartWebTransaction("search_json", w, r)
	defer txn.End()

	results := ctx.app.TypeDB.Search(r.FormValue("q"))
	r.Header["Content-Type"] = []string{"application/json"}
	json.NewEncoder(w).Encode(results)
}
