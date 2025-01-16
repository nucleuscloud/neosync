package utils

import (
	"crypto/sha256"
	"fmt"
	"slices"
)

const (
	CliVersionKey  = "cliVersion"
	CliPlatformKey = "cliPlatform"
	CliCommitKey   = "cliCommit"
)

func ToSha256(input string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(input)))
}

func DedupeSliceOrdered[T comparable](input []T) []T {
	set := map[T]any{}
	output := []T{}
	for _, i := range input {
		if _, exists := set[i]; !exists {
			set[i] = struct{}{}
			output = append(output, i)
		}
	}
	return output
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
