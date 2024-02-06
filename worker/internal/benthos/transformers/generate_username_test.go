package transformers

import (
	"fmt"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateUsername(t *testing.T) {

	res, err := GenerateUsername(maxLength)
	assert.NoError(t, err)

	assert.IsType(t, "", res, "The expected username should have a valid username")
	assert.LessOrEqual(t, int64(len(res)), maxLength, fmt.Sprintf("The city should be less than or equal to the max length. This is the error city:%s", res))
}

func Test_GenerateUsernameShort(t *testing.T) {

	res, err := GenerateUsername(3)
	assert.NoError(t, err)

	assert.IsType(t, "", res, "The expected username should have a valid username")
	assert.LessOrEqual(t, int64(len(res)), int64(3), fmt.Sprintf("The city should be less than or equal to the max length. This is the error city:%s", res))
}

func Test_UsernamelTransformer(t *testing.T) {
	mapping := fmt.Sprintf(`root = generate_username(max_length: %d)`, maxLength)
	ex, err := bloblang.Parse(mapping)

	assert.NoError(t, err, "failed to parse the realistic username transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.IsType(t, "", res, "The expected username should have a valid username")
}
