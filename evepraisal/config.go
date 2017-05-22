package main

import "github.com/spf13/viper"

func init() {
	viper.SetDefault("http_addr", ":8080")
	viper.SetDefault("https_addr", "")
	viper.SetDefault("https_domain-whitelist", []string{"evepraisal.com"})
	viper.SetDefault("https_cert-cache-path", "db/certs")
	viper.SetDefault("crest_baseurl", "https://crest-tq.eveonline.com")
	viper.SetDefault("cache_db", "db/cache")
	viper.SetDefault("appraisal_db", "db/appraisals")
	viper.SetDefault("type_db", "db/static")
	viper.SetDefault("type_static-file", "https://cdn1.eveonline.com/data/sde/tranquility/sde-20170509-TRANQUILITY.zip")
	viper.SetDefault("newrelic_app-name", "Evepraisal")
	viper.SetDefault("newrelic_license-key", "")
	viper.SetDefault("management_addr", "127.0.0.1:8090")
}
