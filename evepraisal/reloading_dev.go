package main

// +build dev

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
	go func() {
		for range watcher.Events {
			log.Println("Detected resource changes, reloading templates")
			evepraisal.MustLoadTemplateFiles()
		}
	}()
}
