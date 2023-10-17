package neosync_plugins

import (
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestProcessUuid(t *testing.T) {

	tests := []struct {
		includeHyphens bool
	}{
		{true},  // checks hyphens
		{false}, // checks no hyphens
	}

	for _, tt := range tests {
		uuidString, err := ProcessUuid(tt.includeHyphens)

		if err != nil {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)

			assert.True(t, isValidUuid(uuidString))

		}
	}

}

// the uuid lib will validate both hyphens and hyphens
func isValidUuid(uuidString string) bool {
	_, err := uuid.Parse(uuidString)
	return err == nil
}

func TestUUIDTransformer(t *testing.T) {
	mapping := `root = this.uuidtransformer(true)`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the uuid transformer")

	res, err := ex.Query("test") // input is ignored here
	assert.NoError(t, err)

	assert.Len(t, res.(string), 36, "UUIDs with hyphens must have 36 characters")
}
