package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	evepraisal "github.com/evepraisal/go-evepraisal"
)

func main() {

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	addr := ":8080"
	server := evepraisal.HTTPServer(addr)
	log.Printf("Starting http server (%s)", addr)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatalf("HTTP server failure: %s", err)
		}
	}()

	stopCtx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	defer server.Shutdown(stopCtx)

	go func() {
		err := evepraisal.FetchDataLoop()
		if err != nil {
			log.Fatalf("Fetch market data failure: %s", err)
		}
	}()

	<-stop
	log.Println("Shutting down")
}
