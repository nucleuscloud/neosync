package transformers

import (
	"fmt"
	"testing"
	"time"

	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GenerateRandomBusinessName(t *testing.T) {
	randomizer := rng.New(1)
	res, err := generateRandomBusinessName(randomizer, nil, maxCharacterLimit)

	assert.NoError(t, err)
	assert.NotEmpty(t, res)
	assert.LessOrEqual(t, int64(len(res)), maxCharacterLimit, "The business name should be less than or equal to the max character limit")
}

func Test_GenerateRandomBusinessName_Random_Seed(t *testing.T) {
	seed := time.Now().UnixNano()
	randomizer := rng.New(seed)
	res, err := generateRandomBusinessName(randomizer, nil, maxCharacterLimit)

	assert.NoError(t, err, "failed with seed", "seed", seed)
	assert.NotEmpty(t, res)
	assert.LessOrEqual(t, int64(len(res)), maxCharacterLimit, "The business name should be less than or equal to the max character limit")
}

func Test_GenerateRandomBusinessName_Clamped(t *testing.T) {
	randomizer := rng.New(1)
	res, err := generateRandomBusinessName(randomizer, shared.Ptr(int64(10)), maxCharacterLimit)

	assert.NoError(t, err)
	assert.NotEmpty(t, res)
	assert.LessOrEqual(t, int64(len(res)), maxCharacterLimit, "The business name should be less than or equal to the max character limit")
	assert.GreaterOrEqual(t, int64(len(res)), int64(10))
}

func Test_GenerateRandomBusinessNameTransformer(t *testing.T) {
	mapping := fmt.Sprintf(`root = generate_business_name(max_length:%d)`, maxCharacterLimit)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the business name transformer")
	assert.NotNil(t, ex)

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	resStr, ok := res.(string)
	require.True(t, ok)

	assert.NotEmpty(t, resStr)
	assert.LessOrEqual(t, int64(len(resStr)), maxCharacterLimit, "output should be less than or equal to max char limit")
}

func Test_GenerateRandomBusinessNameTransformer_NoOptions(t *testing.T) {
	mapping := `root = generate_business_name()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the business name transformer")
	assert.NotNil(t, ex)

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	resStr, ok := res.(string)
	require.True(t, ok)

	assert.NotEmpty(t, resStr)
}
