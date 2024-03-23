package transformers

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GenerateRandomLastName(t *testing.T) {
	randomizer := rand.New(rand.NewSource(1))
	res, err := generateRandomLastName(randomizer, nil, maxCharacterLimit)

	assert.NoError(t, err)
	assert.NotEmpty(t, res)
	assert.LessOrEqual(t, int64(len(res)), maxCharacterLimit, "The last name should be less than or equal to the max character limit")
}

func Test_GenerateRandomLastNameTransformer(t *testing.T) {
	mapping := fmt.Sprintf(`root = generate_last_name(max_length:%d)`, maxCharacterLimit)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the last name transformer")
	assert.NotNil(t, ex)

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	resStr, ok := res.(string)
	require.True(t, ok)

	assert.NotEmpty(t, resStr)
	assert.LessOrEqual(t, int64(len(resStr)), maxCharacterLimit, "The last name should be less than or equal to max char limit")
}
