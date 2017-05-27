package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/evepraisal/go-evepraisal"
	"github.com/evepraisal/go-evepraisal/bolt"
	"github.com/evepraisal/go-evepraisal/crest"
	"github.com/evepraisal/go-evepraisal/management"
	"github.com/evepraisal/go-evepraisal/newrelic"
	"github.com/evepraisal/go-evepraisal/noop"
	"github.com/evepraisal/go-evepraisal/parsers"
	"github.com/evepraisal/go-evepraisal/staticdump"
	"github.com/evepraisal/go-evepraisal/web"
	"github.com/spf13/viper"
	"golang.org/x/crypto/acme/autocert"
)

func appMain() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	log.Println("Starting cache DB")
	cacheDB, err := bolt.NewCacheDB(viper.GetString("cache_db"))
	if err != nil {
		log.Fatalf("Unable to open cacheDB: %s", err)
	}
	defer func() {
		err := cacheDB.Close()
		if err != nil {
			log.Fatalf("Problem closing cacheDB: %s", err)
		}
	}()

	log.Println("Checking for type data")
	if _, err := os.Stat(viper.GetString("type_db")); os.IsNotExist(err) {
		loadTypes()
	} else if err != nil {
		log.Fatalf("Unable to check type db file path: %s", err)
	}

	log.Println("Starting type DB")
	typeDB, err := bolt.NewTypeDB(viper.GetString("type_db"), false)
	if err != nil {
		log.Fatalf("Couldn't start type database: %s", err)
	}
	defer func() {
		err := typeDB.Close()
		if err != nil {
			log.Fatalf("Problem closing typeDB: %s", err)
		}
	}()

	log.Println("Starting price DB")
	priceDB, err := crest.NewPriceDB(cacheDB, viper.GetString("crest_baseurl"))
	if err != nil {
		log.Fatalf("Couldn't start price database")
	}
	defer func() {
		err := priceDB.Close()
		if err != nil {
			log.Fatalf("Problem closing priceDB: %s", err)
		}
	}()

	log.Println("Starting appraisal DB")
	appraisalDB, err := bolt.NewAppraisalDB(viper.GetString("appraisal_db"))
	if err != nil {
		log.Fatalf("Couldn't start appraisal database: %s", err)
	}
	defer func() {
		err := appraisalDB.Close()
		if err != nil {
			log.Fatalf("Problem closing appraisalDB: %s", err)
		}
	}()

	log.Println("Starting txn logger")
	var txnLogger evepraisal.TransactionLogger
	if viper.GetString("newrelic_license-key") == "" {
		log.Println("Using no op transaction logger")
		txnLogger = noop.NewTransactionLogger()
	} else {
		log.Println("Using new relic transaction logger")
		txnLogger, err = newrelic.NewTransactionLogger(viper.GetString("newrelic_app-name"), viper.GetString("newrelic_license-key"))
		if err != nil {
			log.Fatalf("Problem starting transaction logger: %s", err)
		}
	}

	app := &evepraisal.App{
		AppraisalDB:       appraisalDB,
		PriceDB:           priceDB,
		TypeDB:            typeDB,
		CacheDB:           cacheDB,
		TransactionLogger: txnLogger,
		Parser: evepraisal.NewContextMultiParser(
			typeDB,
			[]parsers.Parser{
				parsers.ParseKillmail,
				parsers.ParseEFT,
				parsers.ParseFitting,
				parsers.ParseLootHistory,
				parsers.ParsePI,
				parsers.ParseViewContents,
				parsers.ParseWallet,
				parsers.ParseSurveyScan,
				parsers.ParseContract,
				parsers.ParseAssets,
				parsers.ParseIndustry,
				parsers.ParseCargoScan,
				parsers.ParseDScan,
				parsers.NewContextListingParser(typeDB),
				parsers.NewHeuristicParser(typeDB),
			}),
	}

	app.WebContext = web.NewContext(
		app,
		strings.TrimSuffix(viper.GetString("base-url"), "/"),
		viper.GetString("extra-js"),
		viper.GetString("ad-block"))

	servers := mustStartServers(app.WebContext.HTTPHandler())
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

	startEnvironmentWatchers(app)

	log.Printf("Starting Management HTTP server (%s)", viper.GetString("management_addr"))
	mgmtServer := &http.Server{
		Addr:    viper.GetString("management_addr"),
		Handler: management.HTTPHandler(app),
	}
	defer mgmtServer.Close()
	go func() {
		err := mgmtServer.ListenAndServe()
		if err == http.ErrServerClosed {
			log.Println("Management HTTP server stopped")
		} else if err != nil {
			log.Fatalf("Management HTTP server failure: %s", err)
		}
	}()

	<-stop
	log.Println("Shutting down")
}

func mustStartServers(handler http.Handler) []*http.Server {
	servers := make([]*http.Server, 0)

	if viper.GetString("https_addr") != "" {
		log.Printf("Starting HTTPS server (%s) (%s)", viper.GetString("https_addr"), viper.GetStringSlice("https_domain-whitelist"))

		autocertManager := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(viper.GetStringSlice("https_domain-whitelist")...),
			Cache:      autocert.DirCache(viper.GetString("https_cert-cache-path")),
		}

		server := &http.Server{
			Addr:      viper.GetString("https_addr"),
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

	if viper.GetString("http_addr") != "" {
		log.Printf("Starting HTTP server (%s)", viper.GetString("http_addr"))

		server := &http.Server{
			Addr:    viper.GetString("http_addr"),
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

func loadTypes() {
	types, err := staticdump.LoadTypes(viper.GetString("type_static-cache"), viper.GetString("type_static-file"))
	if err != nil {
		log.Fatalf("Unable to load types from static data: %s", err)
	}

	typeDB, err := bolt.NewTypeDB(viper.GetString("type_db"), true)
	if err != nil {
		log.Fatalf("Couldn't start type database: %s", err)
	}
	defer typeDB.Close()

	for _, t := range types {
		err = typeDB.PutType(t)
		if err != nil {
			log.Fatalf("Cannot insert type: %s", err)
		}
	}
}
