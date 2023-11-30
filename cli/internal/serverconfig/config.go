package serverconfig

import "github.com/spf13/viper"

func GetApiBaseUrl() string {
	baseurl := viper.GetString("NEOSYNC_API_URL")
	if baseurl == "" {
		return "http://localhost:8080"
	}
	return baseurl
}
