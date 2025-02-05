package transformers

import (
	"fmt"
	"testing"
	"time"

	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

var testStringValue = "hello"

func Test_TransformStringPreserveLengthTrue(t *testing.T) {
	res, err := transformString(rng.New(time.Now().UnixNano()), &testStringValue, true, 3, maxCharacterLimit)

	assert.NoError(t, err)
	assert.Equal(t, len(testStringValue), len(*res), "The output string should be as long as the input string")
}

func Test_TransformStringPreserveLengthFalse(t *testing.T) {
	res, err := transformString(rng.New(time.Now().UnixNano()), &testStringValue, false, 3, maxCharacterLimit)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(*res), 3, "The expected value should be greater than or equal to 3")
	assert.LessOrEqual(t, int64(len(*res)), maxCharacterLimit, "The expected value should be less than or equal to the max character limit. ")
}

func Test_TransformStringMaxLength(t *testing.T) {
	res, err := transformString(rng.New(time.Now().UnixNano()), &testStringValue, false, 3, maxCharacterLimit)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(*res), 3, "The expected value should be greater than or equal to 3")
	assert.LessOrEqual(t, int64(len(*res)), maxCharacterLimit, "The expected value should be less than or equal to the max character limit. ")
}

func Test_TransformStringTransformer(t *testing.T) {
	mapping := fmt.Sprintf(`root = transform_string(value:%q,preserve_length:true,min_length:%d,max_length:%d)`, testStringValue, 3, maxCharacterLimit)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the random string transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.NotNil(t, res, "The response shouldn't be nil.")

	resStr, ok := res.(*string)
	if !ok {
		t.Errorf("Expected *string, got %T", res)
		return
	}

	if resStr != nil {
		assert.Equal(t, len(testStringValue), len(*resStr), "Generated string must be the same length as the input string")
		assert.IsType(t, *resStr, "", "The actual value type should be a string")
	} else {
		t.Error("Pointer is nil, expected a valid string pointer")
	}
}

func Test_TransformStringTransformerWithEmptyValue(t *testing.T) {
	nilString := ""
	mapping := fmt.Sprintf(`root = transform_string(value:%q,preserve_length:true,min_length:%d,max_length:%d)`, nilString, 3, maxCharacterLimit)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the email transformer")

	_, err = ex.Query(nil)
	assert.NoError(t, err)
}

func Test_TransformStringTransformer_NoOptions(t *testing.T) {
	mapping := fmt.Sprintf(`root = transform_string(value:%q)`, testStringValue)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the email transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
}
