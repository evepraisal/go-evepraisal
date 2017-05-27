package main

import (
	"fmt"
	"log"
	"os"

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

	log.Println("Config settings")
	for k, v := range viper.AllSettings() {
		log.Printf(" -  %s\t%s", k, v)
	}

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
