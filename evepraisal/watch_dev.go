// +build dev

package main

import (
	"log"
	"os"

	"github.com/dietsche/rfsnotify"
	"github.com/evepraisal/go-evepraisal"
	"gopkg.in/fsnotify.v1"
)

func startEnvironmentWatchers(app *evepraisal.App) {
	watcher, err := rfsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Not able to set up resource watcher: %s", err)
	}

	err = watcher.AddRecursive("web/resources/")
	if err != nil {
		log.Fatalf("Not able to set up resource watcher: %s", err)
	}
	go func() {
		for event := range watcher.Events {
			switch event.Op {
			case fsnotify.Create, fsnotify.Remove, fsnotify.Rename:
				log.Println("Detected new, removed or renamed resources, shutting down")
				os.Exit(1)
			case fsnotify.Write:
				log.Println("Detected resource changes, reloading templates")
				err := app.WebContext.Reload()
				if err != nil {
					log.Printf("Could not reload templates %s", err)
				} else {
					log.Println("Done reloading templates")
				}
			}
		}
	}()
}
