package serverconfig

import "github.com/spf13/viper"

var defaultBaseUrl string = "http://localhost:8080"

func GetApiBaseUrl() string {
	baseurl := viper.GetString("NEOSYNC_API_URL")
	if baseurl == "" {
		return defaultBaseUrl
	}
	return baseurl
}
