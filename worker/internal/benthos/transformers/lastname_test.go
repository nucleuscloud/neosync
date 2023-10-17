package neosync_transformers

import (
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func TestProcessLastName(t *testing.T) {

	tests := []struct {
		fn             string
		preserveLength bool
		expectedLength int
	}{
		{"johnson", false, 0},
		{"smith", true, 5}, // checks preserve length
	}

	for _, tt := range tests {
		res, err := ProcessLastName(tt.fn, tt.preserveLength)

		assert.NoError(t, err)

		if tt.preserveLength {
			assert.Equal(t, tt.expectedLength, len(res))

		} else {
			assert.IsType(t, "", res) // Check if the result is a string
		}

	}

}

func TestLastNameTransformer(t *testing.T) {
	mapping := `root = this.lastnametransformer(true)`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the first name transformer")

	testVal := "johnson"

	res, err := ex.Query(testVal)
	assert.NoError(t, err)

	assert.Len(t, res.(string), len(testVal), "Generated last name must be as long as input last name")
}
