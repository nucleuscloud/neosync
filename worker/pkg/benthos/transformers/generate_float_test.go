package transformers

import (
	"fmt"
	"testing"
	"time"

	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

func Test_GenerateRandomFloat(t *testing.T) {
	type testcase struct {
		randomizer    rng.Rand
		randomizeSign bool
		min           float64
		max           float64
		precision     *int64
		scale         *int64

		floor float64
		ceil  float64
	}
	testcases := []testcase{
		{randomizer: rng.New(time.Now().UnixNano()), randomizeSign: false, min: 0, max: 100, precision: shared.Ptr(int64(7)), floor: 0, ceil: 100},
		{randomizer: rng.New(time.Now().UnixNano()), randomizeSign: false, min: -100, max: 100, precision: shared.Ptr(int64(7)), floor: -100, ceil: 100},
		{randomizer: rng.New(time.Now().UnixNano()), randomizeSign: true, min: 20, max: 40, precision: shared.Ptr(int64(7)), floor: -40, ceil: 40},
		{randomizer: rng.New(time.Now().UnixNano()), randomizeSign: false, min: 12.3, max: 19.2, precision: shared.Ptr(int64(7)), floor: 12.3, ceil: 19.2},
		{randomizer: rng.New(time.Now().UnixNano()), randomizeSign: false, min: 12.3, max: 19.2, precision: shared.Ptr(int64(3)), scale: shared.Ptr(int64(1)), floor: 12.3, ceil: 19.2},
	}

	for _, tc := range testcases {
		t.Run("", func(t *testing.T) {
			output, err := generateRandomFloat64(tc.randomizer, tc.randomizeSign, tc.min, tc.max, tc.precision, tc.scale)
			require.NoError(t, err)
			require.GreaterOrEqual(t, output, tc.floor)
			require.LessOrEqual(t, output, tc.ceil)
		})
	}
}

func Test_GenerateRandomFloat_Randomized_Sign(t *testing.T) {
	type testcase struct {
		min       float64
		max       float64
		precision int64

		negativeFloor float64
		negativeCeil  float64

		positiveFloor float64
		positiveCeil  float64
	}
	testcases := []testcase{
		{min: 20, max: 40, precision: 7, negativeFloor: -40, negativeCeil: -20, positiveFloor: 20, positiveCeil: 40},
		{min: 0, max: 40, precision: 7, negativeFloor: -40, negativeCeil: 0, positiveFloor: 0, positiveCeil: 40},
	}

	for _, tc := range testcases {
		t.Run("", func(t *testing.T) {
			output, err := generateRandomFloat64(rng.New(time.Now().UnixNano()), true, tc.min, tc.max, &tc.precision, nil)
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

func Test_GenerateRandomFloat_Benthos(t *testing.T) {
	minValue := float64(9.2)
	maxValue := float64(9.7)
	randomizeSign := false
	precision := int64(7)
	scale := int64(1)

	mapping := fmt.Sprintf(`root = generate_float64(randomize_sign:%t, min:%f, max:%f, precision: %d, scale: %d)`, randomizeSign, minValue, maxValue, precision, scale)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the generate float transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.GreaterOrEqual(t, res.(float64), minValue, "The result should be greater or equal to the minimum")
	assert.LessOrEqual(t, res.(float64), maxValue, "The result should be less or equal to the maximum")
}

func Test_GenerateRandomFloat_Benthos_NoOptions(t *testing.T) {
	mapping := `root = generate_float64()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the generate float transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
}
