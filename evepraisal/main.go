package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/evepraisal/go-evepraisal"
	"github.com/spf13/viper"
	"github.com/syndtr/goleveldb/leveldb"
)

func main() {

	viper.SetConfigType("toml")
	viper.SetConfigName("evepraisal")
	viper.AddConfigPath("/etc/evepraisal/")
	viper.AddConfigPath(".")
	viper.SetEnvPrefix("evepraisal")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()

	if err != nil {
		switch err.(type) {
		case *os.PathError:
			log.Println("No config file found, using defaults")
		default:
			log.Fatalf("Fatal error config file: %s", err)
		}
	}

	log.Println("Config settings")
	for k, v := range viper.AllSettings() {
		log.Printf(" -  %s\t%s", k, v)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	server := evepraisal.HTTPServer()
	log.Printf("Starting http server (%s)", viper.GetString("web.addr"))

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatalf("HTTP server failure: %s", err)
		}
	}()

	stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer server.Shutdown(stopCtx)
	defer cancel()

	cacheDB, err := leveldb.OpenFile(viper.GetString("cache.dir"), nil)
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
