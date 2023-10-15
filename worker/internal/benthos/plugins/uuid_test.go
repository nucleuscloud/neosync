package neosync_plugins

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUuidlTransformer(t *testing.T) {

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

			if tt.includeHyphens {
				assert.True(t, isValidUUID(uuidString))
			} else {
				assert.True(t, isValidUUID(uuidString))
			}
		}
	}

}

// the uuid lib will validate both hyphens and hyphens
func isValidUUID(uuidString string) bool {
	_, err := uuid.Parse(uuidString)
	return err == nil
}
