package utils

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"

	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
)

func FilterSlice[T any](slice []T, filterFn func(T) bool) []T {
	filteredResults := []T{}
	for _, element := range slice {
		if filterFn(element) {
			filteredResults = append(filteredResults, element)
		}
	}
	return filteredResults
}

func MapSlice[T any, V any](slice []T, fn func(T) V) []V {
	newSlice := make([]V, len(slice))
	for index, element := range slice {
		newSlice[index] = fn(element)
	}
	return newSlice
}

func ToSha256(input string) string {
	h := sha256.New()
	h.Write([]byte(input))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

func GetBearerTokenFromHeader(
	header http.Header,
	key string,
) (string, error) {
	unparsedToken := header.Get(key)
	if unparsedToken == "" {
		return "", nucleuserrors.NewUnauthenticated("must provide valid bearer token")
	}
	pieces := strings.Split(unparsedToken, " ")
	if len(pieces) != 2 {
		return "", nucleuserrors.NewUnauthenticated("token not in proper format")
	}
	if pieces[0] != "Bearer" {
		return "", nucleuserrors.NewUnauthenticated("must provided bearer token")
	}
	token := pieces[1]
	return token, nil
}
