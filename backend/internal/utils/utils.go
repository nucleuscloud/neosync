package utils

import (
	"crypto/sha256"
	"fmt"
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
