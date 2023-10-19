package neosync_transformers

import (
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func TestProcessRandomInt(t *testing.T) {

	tests := []struct {
		i              int64
		preserveLength bool
		intLength      int64
		expectedLength int
	}{
		{67543543, false, 0, 0}, // check base int generation
		{324356, false, 6, 6},   // check int generation with a given length
		{324502453, true, 0, 9}, // check preserveLength of input int
	}

	for _, tt := range tests {
		res, err := ProcessRandomInt(tt.i, tt.preserveLength, tt.intLength)

		assert.NoError(t, err)

		if tt.preserveLength {
			assert.Equal(t, GetIntLength(res), tt.expectedLength)
		}

		if !tt.preserveLength && tt.intLength == 0 {
			assert.Equal(t, GetIntLength(res), 4)
		}

		if !tt.preserveLength && tt.intLength > 0 {
			assert.Equal(t, GetIntLength(res), 6)
		}
	}

}

func TestRandomIntTransformer(t *testing.T) {
	mapping := `root = this.randominttransformer(true, 6)`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the random int transformer")

	testVal := int64(397283)

	res, err := ex.Query(testVal)
	assert.NoError(t, err)

	assert.Equal(t, GetIntLength(testVal), GetIntLength(res.(int64))) // Generated int must be the same length as the input int"
	assert.IsType(t, res, testVal)
}
