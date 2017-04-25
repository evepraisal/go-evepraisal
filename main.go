package main

import (
	"log"

	"github.com/evepraisal/go-evepraisal/parsers"
)

func main() {
	log.Println(parsers.ParseAssets([]string{"Sleeper Data Library\t1.080\tSleeper Components\t\t\t10.80 m3"}))
}
