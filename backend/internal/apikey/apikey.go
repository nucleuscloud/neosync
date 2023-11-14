package apikey

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

const (
	prefix         = "neo"
	accountTokenId = "at"
	v1             = "v1"
	separator      = "_"

	uuidPattern = `[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-4[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}`
)

var (
	v1Prefix = strings.Join([]string{prefix, accountTokenId, v1}, separator)

	v1AccountTokenPattern = fmt.Sprintf(
		`^(%s)%s(%s)%sv([\d+])%s%s$`,
		prefix, separator, accountTokenId, separator, separator, uuidPattern,
	)
	v1AccountTokenRegex = regexp.MustCompile(v1AccountTokenPattern)
)

func NewV1AccountKey() string {
	return v1AccountKey(uuid.NewString())
}

func v1AccountKey(suffix string) string {
	return v1Prefix + separator + suffix
}

func IsValidV1AccountKey(apikey string) bool {
	return v1AccountTokenRegex.MatchString(apikey)
}
