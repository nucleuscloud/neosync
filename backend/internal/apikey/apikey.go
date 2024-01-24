package apikey

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

type ApiKeyType string

const (
	AccountApiKey ApiKeyType = "account"
	WorkerApiKey  ApiKeyType = "worker"
)

const (
	prefix         = "neo"
	accountTokenId = "at"
	workerTokenId  = "wt"
	v1             = "v1"
	separator      = "_"

	uuidPattern = `[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-4[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}`
)

var (
	v1AtPrefix = strings.Join([]string{prefix, accountTokenId, v1}, separator)
	v1WtPrefix = strings.Join([]string{prefix, workerTokenId, v1}, separator)

	v1AccountTokenPattern = fmt.Sprintf(
		`^(%s)%s(%s)%sv([\d+])%s%s$`,
		prefix, separator, accountTokenId, separator, separator, uuidPattern,
	)
	v1AccountTokenRegex = regexp.MustCompile(v1AccountTokenPattern)

	v1WorkerTokenPattern = fmt.Sprintf(
		`^(%s)%s(%s)%sv([\d+])%s%s$`,
		prefix, separator, workerTokenId, separator, separator, uuidPattern,
	)
	v1WorkerTokenRegex = regexp.MustCompile(v1WorkerTokenPattern)
)

func NewV1AccountKey() string {
	return v1AccountKey(uuid.NewString())
}

func v1AccountKey(suffix string) string {
	return v1AtPrefix + separator + suffix
}

func IsValidV1AccountKey(apikey string) bool {
	return v1AccountTokenRegex.MatchString(apikey)
}

func NewV1WorkerKey() string {
	return v1WorkerKey(uuid.NewString())
}

func v1WorkerKey(suffix string) string {
	return v1WtPrefix + separator + suffix
}

func IsValidV1WorkerKey(apiKey string) bool {
	return v1WorkerTokenRegex.MatchString(apiKey)
}
