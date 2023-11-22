package transformers

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateIntPhoneNumberPreserveLengthTrue(t *testing.T) {

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

func Test_IntPhoneNumberTransformerWithValue(t *testing.T) {
	testVal := int64(6183849282)
	mapping := fmt.Sprintf(`root = intphonetransformer(%d,true)`, testVal)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the int64 phone transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	numStr := strconv.FormatInt(testVal, 10)
	resStr := strconv.FormatInt(res.(int64), 10)

	assert.Equal(t, len(resStr), len(numStr), "Generated phone number must be the same length as the input phone number")
	assert.IsType(t, res, testVal)
}

func Test_IntPhoneNumberTransformerWithNoValue(t *testing.T) {
	mapping := `root = intphonetransformer()`

	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the int64 phone transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	resStr := strconv.FormatInt(res.(int64), 10)

	assert.Equal(t, len(resStr), 10, "Generated phone number must be the same length as the input phone number")
	assert.IsType(t, res, int64(2), "Generated phone number must an int64 type")
}

func Test_IntPhoneNumberTransformerWithZeroValue(t *testing.T) {
	mapping := `root = intphonetransformer(0)`

	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the int64 phone transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	resStr := strconv.FormatInt(res.(int64), 10)

	assert.Equal(t, len(resStr), 10, "Generated phone number must be the same length as the input phone number")
	assert.IsType(t, res, int64(2), "Generated phone number must an int64 type")
}
