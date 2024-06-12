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

// checks if the error message is critical
func IsCriticalError(errMsg string) bool {
	// list of known error messages for when max connections are reached
	criticalErrors := []string{
		"violates foreign key constraint",
		"duplicate key value violates unique constraint",
		"duplicate entry",
		"cannot add or update a child row",
		"a foreign key constraint fails",
		"could not identify an equality operator",
		"violates not-null constraint",
		"failed to send message to redis_hash_output",
		"mapping returned invalid key type",
	}

	for _, errStr := range criticalErrors {
		if containsIgnoreCase(errMsg, errStr) {
			return true
		}
	}
	return false
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
