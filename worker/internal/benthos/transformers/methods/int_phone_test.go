package neosync_benthos_transformers_methods

import (
	"strconv"
	"strings"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func TestGenerateIntPhoneNumberPreserveLengthTrue(t *testing.T) {

	pn := int64(618384928322)
	expectedLength := 12

	res, err := GenerateIntPhoneNumberPreserveLength(pn)

	assert.NoError(t, err)
	numStr := strconv.FormatInt(res, 10)
	assert.Equal(t, len(numStr), expectedLength, "The length of the output phone number should be the same as the input phone number")

}

func TestGenerateIntPhoneNumberPreserveLengthFalse(t *testing.T) {

	res, err := GenerateRandomTenDigitIntPhoneNumber()

	numStr := strconv.FormatInt(res, 10)

	assert.NoError(t, err)
	assert.False(t, strings.Contains(numStr, "-"), "The output int phone number should not contain hyphens and may not be the same length as the input")

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
