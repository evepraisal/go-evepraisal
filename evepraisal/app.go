package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/evepraisal/go-evepraisal"
	"github.com/evepraisal/go-evepraisal/bolt"
	"github.com/evepraisal/go-evepraisal/esi"
	"github.com/evepraisal/go-evepraisal/management"
	"github.com/evepraisal/go-evepraisal/parsers"
	"github.com/evepraisal/go-evepraisal/staticdump"
	"github.com/evepraisal/go-evepraisal/typedb"
	"github.com/evepraisal/go-evepraisal/web"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/gregjones/httpcache"
	newrelic "github.com/newrelic/go-agent"
	"github.com/sethgrid/pester"
	"github.com/spf13/viper"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/oauth2"
)

func appMain() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	log.Println("Starting price DB")
	priceDB, err := bolt.NewPriceDB(filepath.Join(viper.GetString("db_path"), "prices"))
	if err != nil {
		log.Fatalf("Couldn't start price database: %s", err)
	}
	defer func() {
		log.Println("Stopping price db")
		derr := priceDB.Close()
		if derr != nil {
			log.Fatalf("Problem closing priceDB: %s", derr)
		}
	}()

	httpCache, err := bolt.NewHTTPCache(filepath.Join(viper.GetString("db_path"), "httpcache"))
	if err != nil {
		log.Fatalf("Couldn't start httpCache: %s", err)
	}
	defer func() {
		log.Println("Stopping http cache")
		derr := httpCache.Close()
		if derr != nil {
			log.Fatalf("Problem closing httpCache: %s", derr)
		}
	}()

	httpClient := pester.New()
	httpClient.Transport = httpcache.NewTransport(httpCache)
	httpClient.Concurrency = 4
	httpClient.Timeout = 30 * time.Second
	httpClient.Backoff = pester.ExponentialJitterBackoff
	httpClient.MaxRetries = 10
	httpClient.LogHook = func(e pester.ErrEntry) { log.Println(httpClient.FormatError(e)) }

	fetcherCtx, fetcherCancel := context.WithCancel(context.Background())
	priceFetcher, err := esi.NewPriceFetcher(fetcherCtx, priceDB, viper.GetString("esi_baseurl"), httpClient)
	if err != nil {
		log.Fatalf("Couldn't start price fetcher: %s", err)
	}
	defer func() {
		log.Println("Stopping price fetcher")
		fetcherCancel()
		// TODO: don't wait forever or fix the bug that causes it to hang
		derr := priceFetcher.Close()
		if derr != nil {
			log.Fatalf("Problem closing priceDB: %s", derr)
		}
	}()

	log.Println("Starting appraisal DB")
	appraisalDB, err := bolt.NewAppraisalDB(filepath.Join(viper.GetString("db_path"), "appraisals"))
	if err != nil {
		log.Fatalf("Couldn't start appraisal database: %s", err)
	}
	defer func() {
		log.Println("Stopping appraisal DB")
		derr := appraisalDB.Close()
		if derr != nil {
			log.Fatalf("Problem closing appraisalDB: %s", derr)
		}
	}()

	app := &evepraisal.App{
		AppraisalDB: appraisalDB,
		PriceDB:     priceDB,
	}

	log.Println("Starting type fetcher")
	staticdumpHTTPClient := pester.New()
	staticdumpHTTPClient.Concurrency = 1
	// TODO: Remove this once CCP figures out the certificate situation with cdn1.eveonline.com
	staticdumpHTTPClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	staticdumpHTTPClient.Timeout = 600 * time.Second
	staticdumpHTTPClient.Backoff = pester.ExponentialJitterBackoff
	staticdumpHTTPClient.MaxRetries = 2
	staticdumpHTTPClient.LogHook = func(e pester.ErrEntry) { log.Println(httpClient.FormatError(e)) }

	if viper.GetString("newrelic_license-key") != "" {
		newRelicConfig := newrelic.NewConfig(viper.GetString("newrelic_app-name"), viper.GetString("newrelic_license-key"))
		newRelicConfig.ErrorCollector.IgnoreStatusCodes = []int{400, 401, 404}
		var newRelicApplication newrelic.Application
		newRelicApplication, err = newrelic.NewApplication(newRelicConfig)
		if err != nil {
			log.Fatalf("Problem configuring new relic: %s", err)
		}

		app.NewRelicApplication = newRelicApplication
		httpClient.Transport = NewRoundTripper(newRelicApplication, httpClient.Transport)
		staticdumpHTTPClient.Transport = NewRoundTripper(newRelicApplication, staticdumpHTTPClient.Transport)
	}

	staticFetcher, err := staticdump.NewStaticFetcher(staticdumpHTTPClient, viper.GetString("db_path"), func(typeDB typedb.TypeDB) {
		oldTypeDB := app.TypeDB
		app.TypeDB = typeDB
		app.Parser = evepraisal.NewContextMultiParser(
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
				parsers.NewContextListingParser(typeDB),
				parsers.ParseIndustry,
				parsers.ParseContract,
				parsers.ParseAssets,
				parsers.ParseCargoScan,
				parsers.ParseDScan,
				parsers.ParseCompare,
				parsers.ParseMoonLedger,
				parsers.ParseMiningLedger,
				parsers.NewHeuristicParser(typeDB),
			})

		if oldTypeDB != nil {
			log.Println("closing old typedb")
			err = oldTypeDB.Close()
			if err != nil {
				log.Println("error closing old typedb: ", err)
			}
			log.Println("closed old typedb")
		}
	})
	if err != nil {
		log.Fatalf("Couldn't start static fetcher: %s", err)
	}
	defer func() {
		log.Println("Stopping static content fetcher")
		derr := staticFetcher.Close()
		if derr != nil {
			log.Fatalf("Problem closing static fetcher: %s", derr)
		}

		if app.TypeDB != nil {
			derr = app.TypeDB.Close()
			if derr != nil {
				log.Fatalf("Problem closing typeDB: %s", derr)
			}
		}
	}()

	webContext := web.NewContext(app)
	webContext.BaseURL = strings.TrimSuffix(viper.GetString("base-url"), "/")
	webContext.ExtraHTMLHeader = viper.GetString("extra-html-header")
	webContext.ExtraJS = viper.GetString("extra-js")
	webContext.AdBlock = viper.GetString("ad-block")
	webContext.ExtraStaticFilePath = viper.GetString("extra-static-file-path")
	if viper.GetString("cookie-auth-key") != "" {
		webContext.CookieStore = sessions.NewCookieStore(
			[]byte(viper.GetString("cookie-auth-key")),
			[]byte(viper.GetString("cookie-encryption-key")))
	} else {
		webContext.CookieStore = sessions.NewCookieStore(securecookie.GenerateRandomKey(32))
	}
	if viper.GetString("sso-client-id") != "" {
		webContext.OauthConfig = &oauth2.Config{
			ClientID:     viper.GetString("sso-client-id"),
			ClientSecret: viper.GetString("sso-client-secret"),
			Scopes:       []string{},
			Endpoint: oauth2.Endpoint{
				AuthURL:  viper.GetString("sso-authorize-url"),
				TokenURL: viper.GetString("sso-token-url"),
			},
			RedirectURL: viper.GetString("base-url") + "/oauthcallback",
		}
		webContext.OauthVerifyURL = viper.GetString("sso-verify-url")
	}

	app.WebContext = webContext

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
		Handler: management.HTTPHandler(app, filepath.Join(viper.GetString("backup_path"), "appraisals.gz")),
	}

	go func() {
		derr := mgmtServer.ListenAndServe()
		if derr == http.ErrServerClosed {
			log.Println("Management HTTP server stopped")
		} else if derr != nil {
			log.Fatalf("Management HTTP server failure: %s", derr)
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
			Cache:      autocert.DirCache(filepath.Join(viper.GetString("db_path"), "certs")),
			Email:      viper.GetString("letsencrypt_email"),
		}

		server := &http.Server{
			ReadHeaderTimeout: 20 * time.Second,
			ReadTimeout:       120 * time.Second,
			WriteTimeout:      60 * time.Second,
			IdleTimeout:       120 * time.Second,
			Addr:              viper.GetString("https_addr"),
			Handler:           handler,
			TLSConfig:         &tls.Config{GetCertificate: autocertManager.GetCertificate},
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

		// Wrap our http handler
		handler = autocertManager.HTTPHandler(handler)
	}

	if viper.GetString("http_addr") != "" {
		log.Printf("Starting HTTP server (%s)", viper.GetString("http_addr"))

		var httpHandler = handler
		if viper.GetBool("http_redirect") {
			log.Println("Redirecting http traffic to https")
			httpHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				target := "https://" + req.Host + req.URL.Path
				if len(req.URL.RawQuery) > 0 {
					target += "?" + req.URL.RawQuery
				}
				http.Redirect(w, req, target, http.StatusTemporaryRedirect)
			})
		}

		server := &http.Server{
			ReadHeaderTimeout: 20 * time.Second,
			ReadTimeout:       120 * time.Second,
			WriteTimeout:      60 * time.Second,
			IdleTimeout:       120 * time.Second,
			Addr:              viper.GetString("http_addr"),
			Handler:           httpHandler,
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

// NewRoundTripper returns an http.RoundTripper that is tooled for use in the app
func NewRoundTripper(newrelicApp newrelic.Application, original http.RoundTripper) http.RoundTripper {
	if original == nil {
		original = http.DefaultTransport
	}

	return roundTripperFunc(func(request *http.Request) (*http.Response, error) {
		txn := newrelicApp.StartTransaction("http", nil, nil)
		segment := newrelic.StartExternalSegment(txn, request)
		response, err := original.RoundTrip(request)
		segment.Response = response
		segment.End()
		txn.End()

		return response, err
	})
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
