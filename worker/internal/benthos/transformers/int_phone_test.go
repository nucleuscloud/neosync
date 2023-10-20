package neosync_transformers

import (
	"strconv"
	"strings"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func TestProcessIntPhoneNumberPreserveLengthTrue(t *testing.T) {

	pn := int64(618384928322)
	expectedLength := 12

	res, err := ProcessIntPhoneNumber(pn, true)

	assert.NoError(t, err)
	numStr := strconv.FormatInt(res, 10)
	assert.Equal(t, len(numStr), expectedLength)

}

func TestProcessIntPhoneNumberPreserveLengthFalse(t *testing.T) {

	pn := int64(618384928322)

	res, err := ProcessIntPhoneNumber(pn, true)

	numStr := strconv.FormatInt(res, 10)

	assert.NoError(t, err)
	assert.False(t, strings.Contains(numStr, "-"))

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
