package shared

import (
	"net/http"

	http_client "github.com/nucleuscloud/neosync/worker/internal/http/client"
	"github.com/spf13/viper"
)

// General workflow metadata struct that is intended to be common across activities
type WorkflowMetadata struct {
	WorkflowId string
	RunId      string
}

// Holds the environment variable name and the connection id that should replace it at runtime when the Sync activity is launched
type BenthosDsn struct {
	EnvVarKey string
	// Neosync Connection Id
	ConnectionId string
}

// Returns the neosync url found in the environment, otherwise defaults to localhost
func GetNeosyncUrl() string {
	neosyncUrl := viper.GetString("NEOSYNC_URL")
	if neosyncUrl == "" {
		return "http://localhost:8080"
	}
	return neosyncUrl
}

// Returns an instance of *http.Client that includes the Neosync API Token if one was found in the environment
func GetNeosyncHttpClient() *http.Client {
	apikey := viper.GetString("NEOSYNC_API_KEY")
	return http_client.NewWithAuth(&apikey)
}

// Generic util method that turns any value into its pointer
func Ptr[T any](val T) *T {
	return &val
}
