package transformers

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/stretchr/testify/assert"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

var testValue = int64(8384928322)

func Test_GenerateIntPhoneNumberPreserveLengthTrue(t *testing.T) {
	res, err := transformInt64PhoneNumber(rng.New(time.Now().UnixNano()), testValue, true)
	assert.NoError(t, err)

	numStr := strconv.FormatInt(*res, 10)
	assert.Equal(t, int64(len(numStr)), transformer_utils.GetInt64Length(testValue), "The length of the output phone number should be the same as the input phone number")
}

func Test_GenerateIntPhoneNumberPreserveLengthFalse(t *testing.T) {
	res, err := transformInt64PhoneNumber(rng.New(time.Now().UnixNano()), testValue, false)
	assert.NoError(t, err)

	numStr := strconv.FormatInt(*res, 10)
	assert.Equal(t, int64(len(numStr)), transformer_utils.GetInt64Length(testValue), "The length of the output phone number should be the same as the input phone number")
}

func Test_GenerateIntPhoneNumberPreserveLengthFunction(t *testing.T) {
	res, err := generateIntPhoneNumberPreserveLength(rng.New(time.Now().UnixNano()), testValue)
	assert.NoError(t, err)

	numStr := strconv.FormatInt(res, 10)
	assert.False(t, strings.Contains(numStr, "-"), "The output int phone number should not contain hyphens and may not be the same length as the input")
}

func Test_IntPhoneNumberTransformer(t *testing.T) {
	mapping := fmt.Sprintf(`root = transform_int64_phone_number(value:%d,preserve_length:true)`, testValue)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the int64 phone transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.NotNil(t, res, "The response shouldn't be nil.")

	resInt, ok := res.(*int64)
	if !ok {
		t.Errorf("Expected *int64, got %T", res)
		return
	}

	numStr := strconv.FormatInt(testValue, 10)

	if resInt != nil {
		resStr := strconv.FormatInt(*resInt, 10)
		assert.Equal(t, len(resStr), len(numStr), "Generated phone number must be the same length as the input phone number")
		assert.IsType(t, *resInt, testValue, "The phone number should be of type int64")
	} else {
		t.Error("Pointer is nil, expected a valid int64 pointer")
	}
}

func Test_TransformIntPhoneTransformerWithEmptyValue(t *testing.T) {
	nilNum := 0
	mapping := fmt.Sprintf(`root = transform_int64_phone_number(value:%d,preserve_length:true)`, nilNum)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the int phone number transformer")

	_, err = ex.Query(nil)
	assert.NoError(t, err)
}

func Test_TransformIntPhoneTransformer_NoOptions(t *testing.T) {
	mapping := fmt.Sprintf(`root = transform_int64_phone_number(value:%d)`, 123123)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the int phone number transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
}
