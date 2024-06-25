package utils

import (
	"net/http"
	"slices"
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

func AllElementsEqual[T comparable](slice []T, value T) bool {
	for _, el := range slice {
		if el != value {
			return false
		}
	}
	return true
}

func AnyElementEqual[T comparable](slice []T, value T) bool {
	for _, el := range slice {
		if el == value {
			return true
		}
	}
	return false
}

func NoElementEqual[T comparable](slice []T, value T) bool {
	for _, el := range slice {
		if el == value {
			return false
		}
	}
	return true
}

func CompareSlices(slice1, slice2 []string) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	for _, ele := range slice1 {
		if !slices.Contains(slice2, ele) {
			return false
		}
	}
	return true
}

func DedupeSlice[T comparable](input []T) []T {
	set := map[T]any{}
	for _, i := range input {
		set[i] = struct{}{}
	}
	output := make([]T, 0, len(set))
	for key := range set {
		output = append(output, key)
	}
	return output
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
