package transformers

import (
	"fmt"
	"testing"
	"time"

	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

func Test_GenerateRandomInt(t *testing.T) {
	type testcase struct {
		randomizeSign bool
		min           int64
		max           int64

		floor int64
		ceil  int64
	}
	testcases := []testcase{
		{randomizeSign: false, min: 0, max: 100, floor: 0, ceil: 100},
		{randomizeSign: false, min: -100, max: 100, floor: -100, ceil: 100},
		{randomizeSign: true, min: 20, max: 40, floor: -40, ceil: 40},
	}

	for _, tc := range testcases {
		t.Run("", func(t *testing.T) {
			output, err := generateRandomInt64(rng.New(time.Now().UnixNano()), tc.randomizeSign, tc.min, tc.max)
			require.NoError(t, err)
			require.GreaterOrEqual(t, output, tc.floor)
			require.LessOrEqual(t, output, tc.ceil)
		})
	}
}

func Test_GenerateRandomInt_Randomized_Range(t *testing.T) {
	type testcase struct {
		min int64
		max int64

		negativeFloor int64
		negativeCeil  int64

		positiveFloor int64
		positiveCeil  int64
	}
	testcases := []testcase{
		{min: 20, max: 40, negativeFloor: -40, negativeCeil: -20, positiveFloor: 20, positiveCeil: 40},
		{min: 0, max: 40, negativeFloor: -40, negativeCeil: 0, positiveFloor: 0, positiveCeil: 40},
	}

	for _, tc := range testcases {
		t.Run("", func(t *testing.T) {
			output, err := generateRandomInt64(rng.New(time.Now().UnixNano()), true, tc.min, tc.max)
			require.NoError(t, err)
			if output > 0 {
				require.GreaterOrEqual(t, output, tc.positiveFloor)
				require.LessOrEqual(t, output, tc.positiveCeil)
			} else {
				require.GreaterOrEqual(t, output, tc.negativeFloor)
				require.LessOrEqual(t, output, tc.negativeCeil)
			}
		})
	}
}

func Test_GenerateRandomInt_Benthos(t *testing.T) {
	minValue := int64(2)
	maxValue := int64(9)
	randomizeSign := false

	mapping := fmt.Sprintf(`root = generate_int64(randomize_sign:%t, min:%d, max:%d)`, randomizeSign, minValue, maxValue)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the random int transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.GreaterOrEqual(t, res, minValue, "The result should be greater or equal to the minimum")
	assert.LessOrEqual(t, res, maxValue, "The result should be less or equal to the maximum")
}

func Test_GenerateRandomInt_Benthos_NoOptions(t *testing.T) {
	mapping := `root = generate_int64()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the random int transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
}
