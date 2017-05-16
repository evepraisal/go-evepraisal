// +build dev

package main

import (
	"log"

	"github.com/dietsche/rfsnotify"
	"github.com/evepraisal/go-evepraisal"
)

// In dev, we want to reload our templates whenever our resources change
func init() {
	watcher, err := rfsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Not able to set up resource watcher: %s", err)
	}

	watcher.AddRecursive("resources/")
	go reloadTemplates(watcher)
}

func reloadTemplates(watcher *rfsnotify.RWatcher) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Error reloading templates: %v", r)
			reloadTemplates(watcher)
		}
	}()

	for range watcher.Events {
		log.Println("Detected resource changes, reloading templates")
		evepraisal.MustLoadTemplateFiles()
		log.Println("Done reloading templates")
	}
}
