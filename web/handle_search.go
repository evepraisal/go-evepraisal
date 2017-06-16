package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/evepraisal/go-evepraisal/typedb"
)

type SearchPage struct {
	Results []typedb.EveType `json:"results"`
}

func (ctx *Context) HandleSearch(w http.ResponseWriter, r *http.Request) {
	results := ctx.App.TypeDB.Search(r.FormValue("q"))
	if len(results) == 1 {
		http.Redirect(w, r, fmt.Sprintf("/item/%d", results[0].ID), http.StatusPermanentRedirect)
	}
	ctx.render(r, w, "search.html", SearchPage{Results: results})
}

func (ctx *Context) HandleSearchJSON(w http.ResponseWriter, r *http.Request) {
	results := ctx.App.TypeDB.Search(r.FormValue("q"))
	r.Header["Content-Type"] = []string{"application/json"}
	json.NewEncoder(w).Encode(results)
}
