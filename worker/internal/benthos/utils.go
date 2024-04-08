package neosync_benthos

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

func BuildBenthosTable(schema, table string) string {
	if schema != "" {
		return fmt.Sprintf("%s.%s", schema, table)
	}
	return table
}

func HashBenthosCacheKey(jobId, runId, table, col string) string {
	return ToSha256(fmt.Sprintf("%s.%s.%s.%s", jobId, runId, table, col))
}

func ToSha256(input string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(input)))
}

// checks if the error message matches a max connections error
func IsMaxConnectionError(errMsg string) bool {
	// list of known error messages for when max connections are reached
	maxConnErrors := []string{
		"too many clients already",
		"remaining connection slots are reserved",
		"maximum number of connections reached",
	}

	for _, errStr := range maxConnErrors {
		if containsIgnoreCase(errMsg, errStr) {
			return true
		}
	}
	return false
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
