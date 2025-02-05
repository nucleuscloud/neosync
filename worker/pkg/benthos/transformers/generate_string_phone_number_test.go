package transformers

import (
	"fmt"
	"testing"
	"time"

	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateStringPhoneNumber(t *testing.T) {
	minValue := int64(9)
	maxValue := int64(14)

	res, err := generateStringPhoneNumber(rng.New(time.Now().UnixNano()), minValue, maxValue)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(res), 9, "Should be greater than 10 characters in length. 9 for the number and 1 for the plus sign.")
	assert.LessOrEqual(t, len(res), 15, "Should be less than 16 characters in length. 15 for the number and 1 for the plus sign.")
}

func Test_GenerateStringPhoneNumberEqualMinMax(t *testing.T) {
	minValue := int64(12)
	maxValue := int64(12)

	res, err := generateStringPhoneNumber(rng.New(time.Now().UnixNano()), minValue, maxValue)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(res), 8, "Should be greater than 9 characters in length. 9 for the number and 1 for the plus sign.")
	assert.LessOrEqual(t, len(res), 15, "Should be less than 16 characters in length. 15 for the number and 1 for the plus sign.")
	assert.Equal(t, int64(len(res)), maxValue)
}

func Test_GenerateStringPhoneNumberShortMax(t *testing.T) {
	minValue := int64(9)
	maxPhoneLimit := 11

	res, err := generateStringPhoneNumber(rng.New(time.Now().UnixNano()), minValue, int64(maxPhoneLimit))

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(res), 8, "Should be greater than 9 characters in length. 9 for the number and 1 for the plus sign.")
	assert.LessOrEqual(t, len(res), maxPhoneLimit, "Should be less than 16 characters in length. 15 for the number and 1 for the plus sign.")
}

func Test_GenerateStringPhoneNumberTransformer(t *testing.T) {
	minValue := int64(10)
	maxValue := int64(13)
	mapping := fmt.Sprintf(`root = generate_string_phone_number(min:%d,max:%d)`, minValue, maxValue)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the phone transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.GreaterOrEqual(t, len(res.(string)), 8, "Should be greater than 9 characters in length. 9 for the number and 1 for the plus sign.")
	assert.LessOrEqual(t, len(res.(string)), 15, "Should be less than 16 characters in length. 15 for the number and 1 for the plus sign.")
}

func Test_GenerateStringPhoneNumberTransformer_NoOptions(t *testing.T) {
	mapping := `root = generate_string_phone_number()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the phone transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
}
