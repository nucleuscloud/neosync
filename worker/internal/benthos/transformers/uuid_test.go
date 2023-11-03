package neosync_transformers

import (
	"strings"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestProcessUuidPreserveHyphhensTrue(t *testing.T) {

	res, err := GenerateUuid(true)

	assert.NoError(t, err)
	assert.True(t, strings.Contains(res, "-"))
	assert.True(t, isValidUuid(res), "The UUID should have the right format and be valid")

}

func TestProcessUuidPreserveHyphhensFalse(t *testing.T) {

	res, err := GenerateUuid(false)

	assert.NoError(t, err)
	assert.True(t, isValidUuid(res), "The UUID should have the right format and be valid")
	assert.False(t, strings.Contains(res, "-"))

}

// the uuid lib will validate both hyphens and hyphens
func isValidUuid(uuidString string) bool {
	_, err := uuid.Parse(uuidString)
	return err == nil
}

func TestUUIDTransformer(t *testing.T) {
	mapping := `root = uuidtransformer(true)`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the uuid transformer")

	res, err := ex.Query("test") // input is ignored here
	assert.NoError(t, err)

	assert.Len(t, res.(string), 36, "UUIDs with hyphens must have 36 characters")
}
