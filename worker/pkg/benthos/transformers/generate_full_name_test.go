package transformers

import (
	"fmt"
	"testing"

	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GenerateRandomFullName(t *testing.T) {
	randomizer := rng.New(1)

	res, err := generateRandomFullName(randomizer, maxCharacterLimit)

	assert.NoError(t, err)
	assert.NotEmpty(t, res)
	assert.LessOrEqual(t, int64(len(res)), maxCharacterLimit, "The name should be less than the max character limit")
}

func Test_generateRandomFullName_MinLength_Error(t *testing.T) {
	randomizer := rng.New(1)

	res, err := generateRandomFullName(randomizer, 1)
	assert.Error(t, err)
	assert.Empty(t, res)
}

func Test_generateRandomFullName_Small(t *testing.T) {
	randomizer := rng.New(1)

	res, err := generateRandomFullName(randomizer, 4)
	assert.Error(t, err)
	assert.Empty(t, res, "cannot generate a full name of length 3 (excluding space) as it must generate a first and last name")

	res, err = generateRandomFullName(randomizer, 5)
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
}

func Test_GenerateRandomFullNameTransformer(t *testing.T) {
	mapping := fmt.Sprintf(`root = generate_full_name(max_length:%d)`, maxCharacterLimit)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the first name transformer")
	assert.NotNil(t, ex)

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	resStr, ok := res.(string)
	require.True(t, ok)

	assert.NotEmpty(t, resStr)
	assert.LessOrEqual(t, int64(len(resStr)), maxCharacterLimit, "output should be less than or equal to max char limit")
}

func Test_GenerateRandomFullNameTransformer_NoOptions(t *testing.T) {
	mapping := `root = generate_full_name()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the first name transformer")
	assert.NotNil(t, ex)

	res, err := ex.Query(nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
}
