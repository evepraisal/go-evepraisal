// +build dev

package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/evepraisal/go-evepraisal"
	"github.com/fsnotify/fsnotify"
)

func startEnvironmentWatchers(app *evepraisal.App) {
	log.Println("startEnvironmentWatchers (dev env)")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("ERROR: Not able to set up resource watcher: %s", err)
		return
	}
	defer watcher.Close()

	err = filepath.Walk("web/resources/",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if strings.HasSuffix(path, ".html") {
				err = watcher.Add(path)
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
