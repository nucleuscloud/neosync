package transformers

import (
	"fmt"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateStringPhoneNumber(t *testing.T) {
	min := int64(9)
	max := int64(14)

	res, err := GenerateStringPhoneNumber(min, max, maxCharacterLimit)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(res), 9, "Should be greater than 10 characters in length. 9 for the number and 1 for the plus sign.")
	assert.LessOrEqual(t, len(res), 15, "Should be less than 16 characters in length. 15 for the number and 1 for the plus sign.")
}

func Test_GenerateStringPhoneNumberEqualMinMax(t *testing.T) {
	min := int64(12)
	max := int64(12)

	res, err := GenerateStringPhoneNumber(min, max, maxCharacterLimit)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(res), 8, "Should be greater than 9 characters in length. 9 for the number and 1 for the plus sign.")
	assert.LessOrEqual(t, len(res), 15, "Should be less than 16 characters in length. 15 for the number and 1 for the plus sign.")
	assert.Equal(t, int64(len(res)), max)
}

func Test_GenerateStringPhoneNumberShortMax(t *testing.T) {
	min := int64(9)
	max := int64(12)
	maxPhoneLimit := 11

	res, err := GenerateStringPhoneNumber(min, max, int64(maxPhoneLimit))

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(res), 8, "Should be greater than 9 characters in length. 9 for the number and 1 for the plus sign.")
	assert.LessOrEqual(t, len(res), maxPhoneLimit, "Should be less than 16 characters in length. 15 for the number and 1 for the plus sign.")
}

func Test_GenerateStringPhoneNumberTransformer(t *testing.T) {
	min := int64(10)
	max := int64(13)
	mapping := fmt.Sprintf(`root = generate_string_phone_number(min:%d,max:%d,max_length:%d)`, min, max, maxCharacterLimit)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the phone transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.GreaterOrEqual(t, len(res.(string)), 8, "Should be greater than 9 characters in length. 9 for the number and 1 for the plus sign.")
	assert.LessOrEqual(t, len(res.(string)), 15, "Should be less than 16 characters in length. 15 for the number and 1 for the plus sign.")
}
