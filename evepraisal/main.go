package main

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/viper"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

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
		}
	}

	buf := bytes.NewBufferString("Config settings\n")
	w := tabwriter.NewWriter(buf, 0, 0, 1, ' ', 0)
	for k, v := range viper.AllSettings() {
		if strings.Contains(k, "key") || strings.Contains(k, "secret") {
			_, err = fmt.Fprintf(w, "\t%s\tMASKED\n", k)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			_, err = fmt.Fprintf(w, "\t%s\t%#v\n", k, v)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	err = w.Flush()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(buf.String())

	// Check for our subcommands
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "restore":
			restoreMain()
		default:
			fmt.Printf("%q is not valid command.\n", os.Args[1])
			os.Exit(2)
		}
	} else {
		// Default to starting the main app
		appMain()
	}
}
