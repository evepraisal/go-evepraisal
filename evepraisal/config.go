package main

import "github.com/spf13/viper"

func init() {
	viper.SetDefault("web.http.addr", ":8080")
	viper.SetDefault("web.https.cert-cache-path", "db/certs")
	viper.SetDefault("crest.baseurl", "https://crest-tq.eveonline.com")
	viper.SetDefault("cache.db", "db/cache")
	viper.SetDefault("appraisal.db", "db/appraisals")
}
