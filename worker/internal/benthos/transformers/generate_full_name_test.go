package transformers

import (
	"fmt"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateRandomFullNamePreserveLengthTrue(t *testing.T) {
	res, err := GenerateRandomFullName(maxCharacterLimit)

	assert.NoError(t, err)
	assert.True(t, len(res) >= 6, "The name should be greater than the min length name")
	assert.True(t, len(res) <= 13, "The name should be less than the max character limit")
	assert.IsType(t, "", res, "The first name should be a string")
}

func Test_GenerateRandomFullNameMaxLengthBetween12And5(t *testing.T) {
	res, err := GenerateRandomFullName(10)

	assert.NoError(t, err)
	assert.True(t, len(res) >= 6, "The name should be greater than the min length name")
	assert.True(t, len(res) <= 10, "The name should be less than the max character limit")
	assert.IsType(t, "", res, "The first name should be a string")
}

func Test_GenerateRandomFullNameMaxLengthLessThan5(t *testing.T) {
	res, err := GenerateRandomFullName(4)
	assert.NoError(t, err)
	assert.Equal(t, len(res), 4, "The name should be greater than the min length name")
	assert.IsType(t, "", res, "The first name should be a string")
}

func Test_GenerateFullNamePreserveLengthFalse(t *testing.T) {
	res, err := GenerateRandomFullName(maxCharacterLimit)

	assert.NoError(t, err)
	assert.True(t, len(res) >= 6, "The name should be greater than the min length name")
	assert.True(t, len(res) <= 13, "The name should be less than the max character limit")
	assert.IsType(t, "", res, "The full name should be a string")
}

func Test_GenerateRandomFullNameTransformer(t *testing.T) {
	mapping := fmt.Sprintf(`root = generate_full_name(max_length:%d)`, maxCharacterLimit)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the first name transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(res.(string)), 2, "The first name should be more than 0 characters")
	assert.LessOrEqual(t, len(res.(string)), 13, "The first name should be less than characters")
	assert.IsType(t, "", res, "The first name should be a string")
}
