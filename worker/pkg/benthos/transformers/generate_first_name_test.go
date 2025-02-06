package transformers

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GenerateRandomFirstName(t *testing.T) {
	randomizer := rand.New(rand.NewSource(1))
	res, err := generateRandomFirstName(randomizer, nil, maxCharacterLimit)

	assert.NoError(t, err)
	assert.NotEmpty(t, res)
	assert.LessOrEqual(t, int64(len(res)), maxCharacterLimit, "The first name should be less than or equal to the max character limit")
}

func Test_GenerateRandomFirstName_Random_Seed(t *testing.T) {
	seed := time.Now().UnixNano()
	randomizer := rand.New(rand.NewSource(seed))
	res, err := generateRandomFirstName(randomizer, nil, maxCharacterLimit)

	assert.NoError(t, err, "failed with seed", "seed", seed)
	assert.NotEmpty(t, res)
	assert.LessOrEqual(t, int64(len(res)), maxCharacterLimit, "The first name should be less than or equal to the max character limit")
}

func Test_GenerateRandomFirstName_Clamped(t *testing.T) {
	randomizer := rand.New(rand.NewSource(1))
	res, err := generateRandomFirstName(randomizer, shared.Ptr(int64(10)), maxCharacterLimit)

	assert.NoError(t, err)
	assert.NotEmpty(t, res)
	assert.LessOrEqual(t, int64(len(res)), maxCharacterLimit, "The first name should be less than or equal to the max character limit")
	assert.GreaterOrEqual(t, int64(len(res)), int64(10))
}

func Test_GenerateRandomFirstNameTransformer(t *testing.T) {
	mapping := fmt.Sprintf(`root = generate_first_name(max_length:%d)`, maxCharacterLimit)
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

func Test_GenerateRandomFirstNameTransformer_NoOptions(t *testing.T) {
	mapping := `root = generate_first_name()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the first name transformer")
	assert.NotNil(t, ex)

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	resStr, ok := res.(string)
	require.True(t, ok)

	assert.NotEmpty(t, resStr)
}
