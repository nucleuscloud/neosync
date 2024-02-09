package transformers

import (
	"fmt"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateRandomLastName(t *testing.T) {

	res, err := GenerateRandomLastName(maxCharacterLimit)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(res), 2, "The last name should be more than 2 characters")
	assert.LessOrEqual(t, int64(len(res)), maxCharacterLimit, "The last name should be less than or equal to the max character limit")
	assert.IsType(t, "", res, "The last name should be a string")
}

func Test_GenerateRandomLastNameMaxlengthLessThan12(t *testing.T) {

	lowMaxCharLimit := int64(6)

	res, err := GenerateRandomLastName(lowMaxCharLimit)

	assert.NoError(t, err)
	assert.Greater(t, len(res), 2, "The last name should be more than 2 characters")
	assert.Equal(t, int64(len(res)), lowMaxCharLimit, "The last name should be less than or equal to the max character limit")
	assert.IsType(t, "", res, "The last name should be a string")

}

func Test_GenerateRandomLastNameTransformer(t *testing.T) {
	mapping := fmt.Sprintf(`root = generate_last_name(max_length:%d)`, maxCharacterLimit)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the last name transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(res.(string)), 2, "The last name should be more than 0 characters")
	assert.Less(t, len(res.(string)), 13, "The last name should be more than 0 characters")
	assert.IsType(t, "", res, "The last name should be a string")
}
