package management

import (
	"encoding/json"
	"expvar"
	"net/http"
	"strconv"

	"github.com/boltdb/bolt"
	"github.com/evepraisal/go-evepraisal"
	boltdb "github.com/evepraisal/go-evepraisal/bolt"
	"github.com/husobee/vestigo"
)

type Context struct {
	App *evepraisal.App
}

func (ctx *Context) HandleRestore(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var appraisal evepraisal.Appraisal
	err := json.NewDecoder(r.Body).Decode(&appraisal)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = ctx.App.AppraisalDB.PutNewAppraisal(&appraisal)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func HTTPHandler(app *evepraisal.App) http.Handler {
	BackupHandleFunc := func(w http.ResponseWriter, req *http.Request) {
		db, ok := app.AppraisalDB.(*boltdb.AppraisalDB)
		if !ok {
			http.Error(w, "backup not supported for this database", http.StatusInternalServerError)
			return
		}
		err := db.DB.View(func(tx *bolt.Tx) error {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Disposition", `attachment; filename="appraisals"`)
			w.Header().Set("Content-Length", strconv.Itoa(int(tx.Size())))
			_, err := tx.WriteTo(w)
			return err
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	ctx := Context{App: app}
	router := vestigo.NewRouter()
	router.Get("/backup/appraisals", BackupHandleFunc)
	router.Post("/restore", ctx.HandleRestore)

	router.Handle("/expvar", expvar.Handler())
	return router
}
