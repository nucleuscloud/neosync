package transformers

import (
	"fmt"
	"testing"
	"time"

	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

var testPhone = "1234567890"

func Test_TransformStringPhoneNumber(t *testing.T) {
	res, err := transformPhoneNumber(rng.New(time.Now().UnixNano()), &testPhone, true, maxCharacterLimit)

	assert.NoError(t, err)
	assert.Equal(t, len(*res), len(testPhone), "The result should be the same length as the test phone")
}

func Test_TransformStringPhoneNumberEqualMinMax(t *testing.T) {
	res, err := transformPhoneNumber(rng.New(time.Now().UnixNano()), &testPhone, false, maxCharacterLimit)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(*res), 8, "Should be greater than 9 characters in length. 9 for the number and 1 for the plus sign.")
	assert.LessOrEqual(t, len(*res), 15, "Should be less than 16 characters in length. 15 for the number and 1 for the plus sign.")
}

func Test_TransformStringPhoneNumberTransformer(t *testing.T) {
	mapping := fmt.Sprintf(`root = transform_phone_number(value:%q,preserve_length:true,max_length:%d)`, testPhone, maxCharacterLimit)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the transform string phone transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	resStr, ok := res.(*string)
	if !ok {
		t.Errorf("Expected *string, got %T", res)
		return
	}

	if resStr != nil {
		assert.Equal(t, len(*resStr), len(testPhone), "The result should be the same length as the test phone")
		assert.IsType(t, *resStr, "", "The actual value type should be a string")
	} else {
		t.Error("Pointer is nil, expected a valid string pointer")
	}
}

func Test_TransformStringPhoneNumberTransformerWithEmptyValue(t *testing.T) {
	nilString := ""
	mapping := fmt.Sprintf(`root = transform_phone_number(value:%q,preserve_length:true,max_length:%d)`, nilString, maxCharacterLimit)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the transform phone transformer")

	_, err = ex.Query(nil)
	assert.NoError(t, err)
}

func Test_TransformStringPhoneNumberTransformer_NoOptions(t *testing.T) {
	mapping := fmt.Sprintf(`root = transform_phone_number(value:%q)`, testPhone)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the transform phone transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
}
