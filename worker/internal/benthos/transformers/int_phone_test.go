package neosync_transformers

import (
	"strconv"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func TestProcessIntPhoneNumber(t *testing.T) {

	tests := []struct {
		pn             int64
		preserveLength bool
		expectedLength int
	}{
		{6183849282, false, 0},   // check base phone number generation
		{618384928322, true, 12}, // checks preserve length
	}

	for _, tt := range tests {
		res, err := ProcessIntPhoneNumber(tt.pn, tt.preserveLength)

		assert.NoError(t, err)

		if tt.preserveLength {
			numStr := strconv.FormatInt(res, 10)
			assert.Equal(t, len(numStr), tt.expectedLength)
		}

		if !tt.preserveLength {
			numStr := strconv.FormatInt(res, 10)
			assert.Equal(t, len(numStr), 10)

		}
	}

}

func TestIntPhoneNumberTransformer(t *testing.T) {
	mapping := `root = this.intphonetransformer(true)`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the int64 phone transformer")

	testVal := int64(6183849282)

	res, err := ex.Query(testVal)
	assert.NoError(t, err)

	numStr := strconv.FormatInt(testVal, 10)
	resStr := strconv.FormatInt(res.(int64), 10)

	assert.Equal(t, len(resStr), len(numStr), "Generated phone number must be the same length as the input phone number")
	assert.IsType(t, res, testVal)
}
