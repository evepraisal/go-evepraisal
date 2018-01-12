package management

import (
	"compress/gzip"
	"encoding/json"
	"expvar"
	"net/http"
	"os"
	"strconv"

	"github.com/boltdb/bolt"
	"github.com/evepraisal/go-evepraisal"
	boltdb "github.com/evepraisal/go-evepraisal/bolt"
	"github.com/husobee/vestigo"
)

// Context is the context that the web management needs
type Context struct {
	App                 *evepraisal.App
	AppraisalBackupPath string
}

// HandleRestore is the handler for /restore, this is only used for partial restores
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

// HandleBackup is the handler for /backup/appraisals
func (ctx *Context) HandleBackup(w http.ResponseWriter, req *http.Request) {
	db, ok := ctx.App.AppraisalDB.(*boltdb.AppraisalDB)
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

// HandleBackupToFile is the handler for /backup/appraisals
func (ctx *Context) HandleBackupToFile(w http.ResponseWriter, req *http.Request) {
	db, ok := ctx.App.AppraisalDB.(*boltdb.AppraisalDB)
	if !ok {
		http.Error(w, "backup not supported for this database", http.StatusInternalServerError)
		return
	}

	f, err := os.OpenFile(ctx.AppraisalBackupPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer f.Close()

	cf, err := gzip.NewWriterLevel(f, gzip.BestCompression)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	err = db.DB.View(func(tx *bolt.Tx) error {
		return tx.Copy(cf)
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// HTTPHandler returns the http.Handler for the web management api
func HTTPHandler(app *evepraisal.App, appraisalBackupPath string) http.Handler {
	ctx := Context{App: app, AppraisalBackupPath: appraisalBackupPath}
	router := vestigo.NewRouter()
	router.Get("/backup/appraisals", ctx.HandleBackup)
	router.Get("/backup-to-file/appraisals", ctx.HandleBackupToFile)
	router.Post("/restore", ctx.HandleRestore)
	router.Handle("/expvar", expvar.Handler())
	return router
}
