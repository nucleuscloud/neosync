package transformers

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
	"github.com/stretchr/testify/assert"
)

var testValue = int64(8384928322)

func Test_GenerateIntPhoneNumberPreserveLengthTrue(t *testing.T) {

	res, err := TransformIntPhoneNumber(testValue, true)
	assert.NoError(t, err)

	numStr := strconv.FormatInt(res, 10)
	assert.Equal(t, int64(len(numStr)), transformer_utils.GetIntLength(testValue), "The length of the output phone number should be the same as the input phone number")

}

func Test_GenerateIntPhoneNumberPreserveLengthFalse(t *testing.T) {

	res, err := TransformIntPhoneNumber(testValue, false)
	assert.NoError(t, err)

	fmt.Println("res", res)

	numStr := strconv.FormatInt(res, 10)
	assert.Equal(t, int64(len(numStr)), transformer_utils.GetIntLength(testValue), "The length of the output phone number should be the same as the input phone number")
}

func Test_GenerateIntPhoneNumberPreserveLengthFunction(t *testing.T) {

	res, err := GenerateIntPhoneNumberPreserveLength(testValue)
	assert.NoError(t, err)

	numStr := strconv.FormatInt(res, 10)
	assert.False(t, strings.Contains(numStr, "-"), "The output int phone number should not contain hyphens and may not be the same length as the input")

}

func Test_IntPhoneNumberTransformer(t *testing.T) {
	mapping := fmt.Sprintf(`root = transform_int64_phone(%d,true)`, testValue)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the int64 phone transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	numStr := strconv.FormatInt(testValue, 10)
	resStr := strconv.FormatInt(res.(int64), 10)

	assert.Equal(t, len(resStr), len(numStr), "Generated phone number must be the same length as the input phone number")
	assert.IsType(t, res, testValue, "The phone number should be of type int64")
}
