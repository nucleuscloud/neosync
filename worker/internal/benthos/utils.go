package neosync_benthos

import (
	"crypto/sha256"
	"fmt"
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
	h := sha256.New()
	h.Write([]byte(input))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}
