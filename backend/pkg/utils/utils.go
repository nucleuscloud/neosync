package utils

import (
	"crypto/sha256"
	"fmt"
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
