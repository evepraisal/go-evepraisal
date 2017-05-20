package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/evepraisal/go-evepraisal/bolt"
	"github.com/evepraisal/go-evepraisal/legacy"
	"github.com/evepraisal/go-evepraisal/staticdump"
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
		if _, err := os.Stat(file); os.IsNotExist(err) {
			restoreCmd.PrintDefaults()
			log.Fatalf("File path does not exist: %s", file)
		} else if err != nil {
			restoreCmd.PrintDefaults()
			log.Fatalf("Error checking file: %s", file)
		}
	}

	typeDB, err := staticdump.NewTypeDB("db/static", "https://cdn1.eveonline.com/data/sde/tranquility/sde-20170509-TRANQUILITY.zip", true)
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

	for _, filename := range filenames {
		err := legacy.RestoreLegacyFile(appraisalDB, typeDB, filename)
		if err != nil {
			log.Fatalf("Error while importing legacy file: %s", err)
		}
	}
}
