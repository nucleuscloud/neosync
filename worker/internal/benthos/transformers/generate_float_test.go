package transformers

import (
	"fmt"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GenerateRandomFloat(t *testing.T) {
	type testcase struct {
		randomizeSign bool
		min           float64
		max           float64
		precision     int64

		floor float64
		ceil  float64
	}
	testcases := []testcase{
		{randomizeSign: false, min: 0, max: 100, precision: 7, floor: 0, ceil: 100},
		{randomizeSign: false, min: -100, max: 100, precision: 7, floor: -100, ceil: 100},
		{randomizeSign: true, min: 20, max: 40, precision: 7, floor: -40, ceil: 40},
		{randomizeSign: false, min: 12.3, max: 19.2, precision: 7, floor: 12.3, ceil: 19.2},
		{randomizeSign: false, min: 12.3, max: 19.2, precision: 3, floor: 12.3, ceil: 19.2},
	}

	for _, tc := range testcases {
		t.Run("", func(t *testing.T) {
			output, err := generateRandomFloat64(tc.randomizeSign, tc.min, tc.max, tc.precision)
			require.NoError(t, err)
			require.GreaterOrEqual(t, output, tc.floor)
			require.LessOrEqual(t, output, tc.ceil)
		})
	}
}

func Test_GenerateRandomFloat_Randomized_Range(t *testing.T) {
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
			output, err := generateRandomFloat64(true, tc.min, tc.max, tc.precision)
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
	min := float64(9.2)
	max := float64(9.7)
	randomizeSign := false
	precision := int64(7)

	mapping := fmt.Sprintf(`root = generate_float64(randomize_sign:%t, min:%f, max:%f, precision: %d)`, randomizeSign, min, max, precision)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the generate float transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.GreaterOrEqual(t, res.(float64), min, "The result should be greater or equal to the minimum")
	assert.LessOrEqual(t, res.(float64), max, "The result should be less or equal to the maximum")
}
