package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"text/tabwriter"

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
		}
	}

	buf := bytes.NewBufferString("Config settings\n")
	w := tabwriter.NewWriter(buf, 0, 0, 1, ' ', 0)
	for k, v := range viper.AllSettings() {
		fmt.Fprintf(w, "\t%s\t%#v\n", k, v)
	}
	w.Flush()
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
