package main

import "github.com/spf13/viper"

func init() {
	viper.SetDefault("web", map[string]interface{}{
		"http": map[string]interface{}{
			"addr": ":8080",
		},
		"https": map[string]interface{}{
			"addr":             "",
			"domain-whitelist": []string{"evepraisal.com"},
			"cert-cache-path":  "db/certs",
		},
	})
	viper.SetDefault("crest.baseurl", "https://crest-tq.eveonline.com")
	viper.SetDefault("cache.db", "db/cache")
	viper.SetDefault("appraisal.db", "db/appraisals")
	viper.SetDefault("newrelic.app-name", "Evepraisal")
	viper.SetDefault("type.db", "db/static")
	viper.SetDefault("type.static-file", "https://cdn1.eveonline.com/data/sde/tranquility/sde-20170509-TRANQUILITY.zip")
	viper.SetDefault("management.http.addr", "127.0.0.1:8090")
}
