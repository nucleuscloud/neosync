package shared

import (
	"net/http"

	http_client "github.com/nucleuscloud/neosync/worker/internal/http/client"
	"github.com/spf13/viper"
)

type WorkflowMetadata struct {
	WorkflowId string
	RunId      string
}

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

func Ptr[T any](val T) *T {
	return &val
}
