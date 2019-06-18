package main

import "github.com/spf13/viper"

func init() {
	viper.SetDefault("base-url", "http://127.0.0.1:8080")
	viper.SetDefault("http_addr", ":8080")
	viper.SetDefault("http_redirect", false)
	viper.SetDefault("https_addr", "")
	viper.SetDefault("https_domain-whitelist", []string{"evepraisal.com"})
	viper.SetDefault("letsencrypt_email", "")
	viper.SetDefault("db_path", "db/")
	viper.SetDefault("backup_path", "db/backups/")
	viper.SetDefault("esi_baseurl", "https://esi.evetech.net/latest")
	viper.SetDefault("newrelic_app-name", "Evepraisal")
	viper.SetDefault("newrelic_license-key", "")
	viper.SetDefault("management_addr", "127.0.0.1:8090")
	viper.SetDefault("extra-html-header", "")
	viper.SetDefault("extra-js", "")
	viper.SetDefault("ad-block", "")
	viper.SetDefault("extra-static-file-path", "")
	viper.SetDefault("sso-authorize-url", "https://login.eveonline.com/oauth/authorize")
	viper.SetDefault("sso-token-url", "https://login.eveonline.com/oauth/token")
	viper.SetDefault("sso-verify-url", "https://login.eveonline.com/oauth/verify")
	viper.SetDefault("sso-verify-url", "https://login.eveonline.com/oauth/verify")

}
