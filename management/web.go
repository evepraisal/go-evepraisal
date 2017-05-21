package management

import (
	"encoding/json"
	"expvar"
	"net/http"

	"github.com/evepraisal/go-evepraisal"
	"github.com/husobee/vestigo"
)

type Context struct {
	app *evepraisal.App
}

func (ctx *Context) HandleRestore(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var appraisal evepraisal.Appraisal
	err := json.NewDecoder(r.Body).Decode(&appraisal)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = ctx.app.AppraisalDB.PutNewAppraisal(&appraisal)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func HTTPHandler(app *evepraisal.App) http.Handler {
	ctx := Context{app: app}
	router := vestigo.NewRouter()
	// router.Get("/backup", )
	router.Post("/restore", ctx.HandleRestore)

	router.Handle("/expvar", expvar.Handler())
	return router
}
