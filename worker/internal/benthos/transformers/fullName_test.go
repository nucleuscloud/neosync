package neosync_transformers

import (
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func TestProcessFulltName(t *testing.T) {

	tests := []struct {
		fn             string
		preserveLength bool
		expectedLength int
	}{
		{"frank johnson", false, 0},
		{"evis drenova", true, 12}, // checks preserve length
	}

	for _, tt := range tests {
		res, err := ProcessFullName(tt.fn, tt.preserveLength)

		assert.NoError(t, err)

		if tt.preserveLength {
			assert.Equal(t, tt.expectedLength, len(res))

		} else {
			assert.IsType(t, "", res) // Check if the result is a string
		}

	}

}

func TestFullNameTransformer(t *testing.T) {
	mapping := `root = this.fullnametransformer(true)`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the full name transformer")

	testVal := "john smith"

	res, err := ex.Query(testVal)
	assert.NoError(t, err)

	assert.Len(t, res.(string), len(testVal), "Generated full name must be as long as input full name")
}
