package main

import "github.com/spf13/viper"

func init() {
	viper.SetDefault("base-url", "http://127.0.0.1:8080")
	viper.SetDefault("http_addr", ":8080")
	viper.SetDefault("https_addr", "")
	viper.SetDefault("https_domain-whitelist", []string{"evepraisal.com"})
	viper.SetDefault("db_path", "db/")
	viper.SetDefault("esi_baseurl", "https://esi.tech.ccp.is/latest")
	viper.SetDefault("newrelic_app-name", "Evepraisal")
	viper.SetDefault("newrelic_license-key", "")
	viper.SetDefault("management_addr", "127.0.0.1:8090")
	viper.SetDefault("extra-js", "")
	viper.SetDefault("ad-block", "")
}
