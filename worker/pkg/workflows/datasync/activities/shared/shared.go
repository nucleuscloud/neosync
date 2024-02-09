package shared

import (
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

func Ptr[T any](val T) *T {
	return &val
}
