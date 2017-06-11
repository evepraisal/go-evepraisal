package staticdump

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/evepraisal/go-evepraisal/bolt"
	"github.com/evepraisal/go-evepraisal/typedb"
)

type StaticFetcher struct {
	dbPath   string
	callback func(typeDB typedb.TypeDB)

	stop chan bool
	wg   *sync.WaitGroup
}

func NewStaticFetcher(dbPath string, callback func(typeDB typedb.TypeDB)) (*StaticFetcher, error) {
	fetcher := &StaticFetcher{
		dbPath:   dbPath,
		callback: callback,

		stop: make(chan bool),
		wg:   &sync.WaitGroup{},
	}

	err := fetcher.RunOnce()
	if err != nil {
		return nil, err
	}

	fetcher.wg.Add(1)
	go func() {
		defer fetcher.wg.Done()
		for {

			select {
			case <-time.After(6 * time.Hour):
			case <-fetcher.stop:
				return
			}

			err := fetcher.RunOnce()
			if err != nil {
				log.Printf("WARNING: Fetcher failed to run: %s", err)
			}

			log.Println("FETCHER RAN")
		}
	}()

	return fetcher, nil
}

func (f *StaticFetcher) RunOnce() error {
	staticDumpURL := MustFindLastStaticDumpURL()
	staticDumpURLBase := filepath.Base(staticDumpURL)
	typedbPath := filepath.Join(f.dbPath, "types-"+strings.TrimSuffix(staticDumpURLBase, filepath.Ext(staticDumpURLBase)))
	if _, err := os.Stat(typedbPath); os.IsNotExist(err) {
		f.loadTypes(typedbPath, staticDumpURL)
	} else if err != nil {
		return err
	}

	log.Println("Done loading types")

	typeDB, err := bolt.NewTypeDB(typedbPath, false)
	if err != nil {
		return err
	}

	f.callback(typeDB)
	return nil
}

func (f *StaticFetcher) Close() error {
	close(f.stop)
	f.wg.Wait()
	return nil
}

func (f *StaticFetcher) loadTypes(staticCacheFile string, staticDumpURL string) {
	types, err := LoadTypes(staticCacheFile+".zip", staticDumpURL)
	if err != nil {
		log.Fatalf("Unable to load types from static data: %s", err)
	}

	typeDB, err := bolt.NewTypeDB(staticCacheFile, true)
	if err != nil {
		log.Fatalf("Couldn't start type database (write mode): %s", err)
	}
	finished := false
	defer func() {
		if finished == true {
			typeDB.Close()
		} else {
			log.Println("Deleting new typedb because it was stopped before finishing")
			err := typeDB.Delete()
			if err != nil {
				log.Print("Error deleting old typedb: %s", err)
			}
		}
	}()

	for _, t := range types {
		select {
		case <-f.stop:
			return
		}

		err = typeDB.PutType(t)
		if err != nil {
			log.Fatalf("Cannot insert type: %s", err)
		}
	}
	finished = true
}
