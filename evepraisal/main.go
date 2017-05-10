package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	evepraisal "github.com/evepraisal/go-evepraisal"
	"github.com/syndtr/goleveldb/leveldb"
)

func main() {

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// TODO: configurable
	addr := ":8080"
	server := evepraisal.HTTPServer(addr)
	log.Printf("Starting http server (%s)", addr)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatalf("HTTP server failure: %s", err)
		}
	}()

	stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer server.Shutdown(stopCtx)
	defer cancel()

	// TODO: configurable
	cacheDB, err := leveldb.OpenFile("db/cache", nil)
	if err != nil {
		log.Fatalf("Unable to open cache leveldb: %s", err)
	}
	defer cacheDB.Close()

	go func() {
		err := evepraisal.FetchDataLoop(cacheDB)
		if err != nil {
			log.Fatalf("Fetch market data failure: %s", err)
		}
	}()

	<-stop
	log.Println("Shutting down")
}
