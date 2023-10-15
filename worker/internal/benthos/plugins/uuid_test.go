package neosync_plugins

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUuidTransformer(t *testing.T) {

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
