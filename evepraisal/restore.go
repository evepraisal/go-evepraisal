package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/evepraisal/go-evepraisal"
	"github.com/evepraisal/go-evepraisal/legacy"
	"github.com/evepraisal/go-evepraisal/staticdump"
	"github.com/evepraisal/go-evepraisal/typedb"
	"github.com/sethgrid/pester"
	"github.com/spf13/viper"
)

func restoreMain() {
	restoreCmd := flag.NewFlagSet("restore", flag.ExitOnError)
	filenamesStr := restoreCmd.String("files", "", "comma-separated filenames to import data from")
	err := restoreCmd.Parse(os.Args[2:])
	if err != nil || restoreCmd.Parsed() == false {
		restoreCmd.PrintDefaults()
		os.Exit(2)
	}

	if *filenamesStr == "" {
		restoreCmd.PrintDefaults()
		log.Fatalln("The -filenames option is required")
	}

	filenames := strings.Split(*filenamesStr, ",")
	for _, file := range filenames {
		if _, err = os.Stat(file); os.IsNotExist(err) {
			restoreCmd.PrintDefaults()
			log.Fatalf("File path does not exist: %s", file)
		} else if err != nil {
			restoreCmd.PrintDefaults()
			log.Fatalf("Error checking file: %s", file)
		}
	}

	var typeDB typedb.TypeDB

	staticdumpHTTPClient := pester.New()
	staticdumpHTTPClient.Concurrency = 1
	staticdumpHTTPClient.Timeout = 5 * time.Minute
	staticdumpHTTPClient.Backoff = pester.ExponentialJitterBackoff
	staticdumpHTTPClient.MaxRetries = 10
	staticFetcher, err := staticdump.NewStaticFetcher(staticdumpHTTPClient, viper.GetString("db_path"), func(newTypeDB typedb.TypeDB) {
		log.Println("Got new typedb", newTypeDB)
		typeDB = newTypeDB
	})
	log.Println("TypeDB", typeDB)
	if err != nil {
		log.Fatalf("Couldn't start static fetcher: %s", err)
	}
	defer func() {
		err := staticFetcher.Close()
		if err != nil {
			log.Fatalf("Problem closing static fetcher: %s", err)
		}

		if typeDB != nil {
			err = typeDB.Close()
			if err != nil {
				log.Fatalf("Problem closing typeDB: %s", err)
			}
		}
	}()

	saver := func(appraisal *evepraisal.Appraisal) error {
		var (
			buf bytes.Buffer
			req *http.Request
			err error
		)
		err = json.NewEncoder(&buf).Encode(appraisal)
		if err != nil {
			return err
		}

		req, err = http.NewRequest("POST", "http://"+viper.GetString("management_addr")+"/restore", &buf)
		if err != nil {
			return err
		}

		req.Header.Add("Content-Type", "application/json")

		var resp *http.Response
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			var body []byte
			body, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			log.Printf("ERROR: %s: %s", resp.Status, string(body))
		}

		return resp.Body.Close()
	}

	for _, filename := range filenames {
		log.Printf("Start restoring: %s", filename)
		err := legacy.RestoreLegacyFile(saver, typeDB, filename)
		if err != nil {
			log.Fatalf("Error while importing legacy file: %s", err)
		}
		log.Printf("Done restoring: %s", filename)
	}
}
