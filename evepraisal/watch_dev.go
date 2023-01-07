//go:build dev
// +build dev

package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/evepraisal/go-evepraisal"
	"github.com/radovskyb/watcher"
)

func startEnvironmentWatchers(app *evepraisal.App) {
	log.Println("startEnvironmentWatchers (dev env)")
	w := watcher.New()
	defer w.Close()

	go func() {
		for {
			select {
			case <-w.Closed:
				return
			case event, ok := <-w.Event:
				if !ok {
					return
				}
				log.Printf("%s %s", event.Name, event.Op)
				switch event.Op {
				case watcher.Create, watcher.Remove, watcher.Rename:
					log.Println("Detected new, removed or renamed resources, shutting down")
					os.Exit(1)
				case watcher.Write:
					log.Println("Detected resource changes, reloading templates")
					err := app.WebContext.Reload()
					if err != nil {
						log.Printf("Could not reload templates %s", err)
					} else {
						log.Println("Done reloading templates")
					}
				}
			case err, ok := <-w.Error:
				if !ok {
					return
				}
				log.Printf("ERROR: Watching for fs watcher: %s", err)
			}
		}
	}()

	err := filepath.Walk("web/resources/",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if strings.HasSuffix(path, ".html") {
				err = w.Add(path)
				if err != nil {
					return err
				}
			}
			return nil
		})
	if err != nil {
		log.Printf("ERROR: Not able to set up resource watcher: %s", err)
	}

	go func() {
		if err := w.Start(time.Millisecond * 100); err != nil {
			log.Fatalln(err)
		}
	}()
}
