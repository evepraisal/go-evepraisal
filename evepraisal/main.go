package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"golang.org/x/crypto/acme/autocert"

	"github.com/evepraisal/go-evepraisal"
	"github.com/evepraisal/go-evepraisal/bolt"
	"github.com/evepraisal/go-evepraisal/crest"
	"github.com/evepraisal/go-evepraisal/parsers"
	"github.com/evepraisal/go-evepraisal/staticdump"
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

	priceDB, err := crest.NewPriceDB(cacheDB, viper.GetString("crest.baseurl"))
	if err != nil {
		log.Fatalf("Couldn't start price database")
	}
	defer func() {
		err := priceDB.Close()
		if err != nil {
			log.Fatalf("Problem closing priceDB: %s", err)
		}
	}()

	err = os.MkdirAll("db/static", 0700)
	if err != nil {
		log.Fatalf("Unable to create static data dir: %s", err)
	}
	typeDB, err := staticdump.NewTypeDB("db/static", "https://cdn1.eveonline.com/data/sde/tranquility/sde-20170509-TRANQUILITY.zip")
	if err != nil {
		log.Fatalf("Couldn't start type database: %s", err)
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

	servers := mustStartServers(evepraisal.HTTPHandler(app))
	if err != nil {
		log.Fatalf("Problem starting https server: %s", err)
	}

	for _, server := range servers {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer server.Shutdown(stopCtx)
		go func() {
			time.Sleep(10 * time.Second)
			cancel()
		}()
	}

	<-stop
	log.Println("Shutting down")
}

func mustStartServers(handler http.Handler) []*http.Server {
	servers := make([]*http.Server, 0)

	if viper.GetString("web.https.addr") != "" {
		log.Printf("Starting HTTPS server (%s) (%s)", viper.GetString("web.https.addr"), viper.GetStringSlice("web.https.domain-whitelist"))

		autocertManager := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(viper.GetStringSlice("web.https.domain-whitelist")...),
			Cache:      autocert.DirCache(viper.GetString("web.https.cert-cache-path")),
		}

		server := &http.Server{
			Addr:      viper.GetString("web.https.addr"),
			Handler:   handler,
			TLSConfig: &tls.Config{GetCertificate: autocertManager.GetCertificate},
		}
		servers = append(servers, server)

		go func() {
			err := server.ListenAndServeTLS("", "")
			if err == http.ErrServerClosed {
				log.Println("HTTPS server stopped")
			} else if err != nil {
				log.Fatalf("HTTPS server failure: %s", err)
			}
		}()
		time.Sleep(1 * time.Second)
	}

	if viper.GetString("web.http.addr") != "" {
		log.Printf("Starting HTTP server (%s)", viper.GetString("web.http.addr"))

		server := &http.Server{
			Addr:    viper.GetString("web.http.addr"),
			Handler: handler,
		}
		servers = append(servers, server)

		go func() {
			err := server.ListenAndServe()
			if err == http.ErrServerClosed {
				log.Println("HTTP server stopped")
			} else if err != nil {
				log.Fatalf("HTTP server failure: %s", err)
			}
		}()
	}

	return servers
}
