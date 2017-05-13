package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/evepraisal/go-evepraisal"
	"github.com/evepraisal/go-evepraisal/bolt"
	"github.com/evepraisal/go-evepraisal/crest"
	"github.com/evepraisal/go-evepraisal/parsers"
	"github.com/spf13/viper"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

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

	cacheDB, err := bolt.NewCacheDB(viper.GetString("cache.db"))
	if err != nil {
		log.Fatalf("Unable to open cacheDB: %s", err)
	}
	defer func() {
		err := cacheDB.Close()
		if err != nil {
			log.Fatalf("Problem closing cacheDB: %s", err)
		}
	}()

	priceDB, err := crest.NewPriceDB(cacheDB)
	if err != nil {
		log.Fatalf("Couldn't start price database")
	}
	defer func() {
		err := priceDB.Close()
		if err != nil {
			log.Fatalf("Problem closing priceDB: %s", err)
		}
	}()

	typeDB, err := crest.NewTypeDB(cacheDB)
	if err != nil {
		log.Fatalf("Couldn't start type database")
	}
	defer func() {
		err := typeDB.Close()
		if err != nil {
			log.Fatalf("Problem closing typeDB: %s", err)
		}
	}()

	appraisalDB, err := bolt.NewAppraisalDB(viper.GetString("appraisal.db"))
	if err != nil {
		log.Fatalf("Couldn't start appraisal database")
	}
	defer func() {
		err := appraisalDB.Close()
		if err != nil {
			log.Fatalf("Problem closing appraisalDB: %s", err)
		}
	}()

	app := &evepraisal.App{
		AppraisalDB: appraisalDB,
		PriceDB:     priceDB,
		TypeDB:      typeDB,
		CacheDB:     cacheDB,
		Parser:      parsers.AllParser,
	}

	log.Printf("Starting HTTP server (%s)", viper.GetString("web.addr"))
	server := evepraisal.HTTPServer(app)
	go func() {
		err := server.ListenAndServe()
		if err == http.ErrServerClosed {
			log.Println("HTTP server stopped")
		} else if err != nil {
			log.Fatalf("HTTP server failure: %s %T", err, err)
		}
	}()
	stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer server.Shutdown(stopCtx)
	defer cancel()

	<-stop
	log.Println("Shutting down")
}
