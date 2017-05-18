// +build dev

package main

import (
	"log"

	"github.com/dietsche/rfsnotify"
	"github.com/evepraisal/go-evepraisal"
)

func startEnvironmentWatchers(app *evepraisal.App) {
	watcher, err := rfsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Not able to set up resource watcher: %s", err)
	}

	watcher.AddRecursive("resources/")
	go func() {
		for range watcher.Events {
			log.Println("Detected resource changes, reloading templates")
			err := app.LoadTemplates()
			if err != nil {
				log.Printf("Could not reload templates %s", err)
			} else {
				log.Println("Done reloading templates")
			}
		}
	}()
}
