package neosync_transformers

import (
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func TestProcessFirstName(t *testing.T) {

	tests := []struct {
		fn             string
		preserveLength bool
		expectedLength int
	}{
		{"frank", false, 0},
		{"evis", true, 4}, // checks preserve length
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

func TestFirstNameTransformer(t *testing.T) {
	mapping := `root = this.firstnametransformer(true)`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the first name transformer")

	testVal := "evis"

	res, err := ex.Query(testVal)
	assert.NoError(t, err)

	assert.Len(t, res.(string), len(testVal), "Generated first name must be as long as input first name")
}
