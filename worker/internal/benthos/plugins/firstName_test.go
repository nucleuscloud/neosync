package neosync_plugins

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFirstNameTransformer(t *testing.T) {

	tests := []struct {
		fn             string
		preserveLength bool
		expectedLength int
	}{
		{"frank", false, 0}, // checks preserve domain
		{"evis", true, 4},   // checks preserve length
	}

	for _, tt := range tests {
		res, err := ProcessFirstName(tt.fn, tt.preserveLength)

		assert.NoError(t, err)

		if tt.preserveLength {
			assert.Equal(t, tt.expectedLength, len(res))

		} else {
			assert.IsType(t, "", res) // Check if the result is a string
		}

	}

}
