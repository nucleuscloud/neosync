package transformers

import (
	"fmt"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateRandomFloatPositiveRange(t *testing.T) {

	min := float64(12.3)
	max := float64(19.2)
	precision := int64(7)

	res, err := GenerateRandomFloat64(false, min, max, precision)
	assert.NoError(t, err)

	assert.GreaterOrEqual(t, res, min, "The result should be greater or equal to the minimum")
	assert.LessOrEqual(t, res, max, "The result should be less or equal to the maximum")
}

func Test_GenerateRandomFloatNegativeRange(t *testing.T) {

	min := float64(-12.3)
	max := float64(-19.2)
	precision := int64(7)

	res, err := GenerateRandomFloat64(false, min, max, precision)
	assert.NoError(t, err)

	// swapped because negative min number is the max
	assert.GreaterOrEqual(t, res, max, "The result should be greater or equal to the minimum")
	assert.LessOrEqual(t, res, min, "The result should be less or equal to the maximum")
}

func Test_GenerateRandomFloatNegativetoPositiveRange(t *testing.T) {

	min := float64(-12.3)
	max := float64(19.2)
	precision := int64(3)

	res, err := GenerateRandomFloat64(false, min, max, precision)
	assert.NoError(t, err)

	assert.GreaterOrEqual(t, res, min, "The result should be greater or equal to the minimum")
	assert.LessOrEqual(t, res, max, "The result should be less or equal to the maximum")
}

func Test_GenerateRandomFloatRandomizePositive(t *testing.T) {

	min := float64(12.3)
	max := float64(19.2)
	precision := int64(7)

	res, err := GenerateRandomFloat64(true, min, max, precision)
	assert.NoError(t, err)

	if !transformer_utils.IsNegativeFloat64(res) {
		// res is positive
		assert.GreaterOrEqual(t, res, min, "The result should be greater or equal to the minimum")
		assert.LessOrEqual(t, res, max, "The result should be less or equal to the maximum")

	} else {
		// res is negative
		assert.GreaterOrEqual(t, res, -max, "The result should be greater or equal to the minimum")
		assert.LessOrEqual(t, res, -min, "The result should be less or equal to the maximum")
	}
}

func Test_GenerateRandomFloatRandomizeNegative(t *testing.T) {

	min := float64(-12.3)
	max := float64(-19.2)
	precision := int64(7)

	res, err := GenerateRandomFloat64(true, min, max, precision)
	assert.NoError(t, err)

	if !transformer_utils.IsNegativeFloat64(res) {
		// res is positive
		assert.GreaterOrEqual(t, res, -min, "The result should be greater or equal to the minimum")
		assert.LessOrEqual(t, res, -max, "The result should be less or equal to the maximum")
	} else {
		// res is negative
		assert.GreaterOrEqual(t, res, max, "The result should be greater or equal to the minimum")
		assert.LessOrEqual(t, res, min, "The result should be less or equal to the maximum")
	}
}

func Test_GenerateRandomFloatRandomizeNegativeToPositive(t *testing.T) {

	min := float64(-12.3)
	max := float64(19.2)
	precision := int64(7)

	res, err := GenerateRandomFloat64(true, min, max, precision)
	assert.NoError(t, err)

	if !transformer_utils.IsNegativeFloat64(res) {
		// res is positive
		assert.GreaterOrEqual(t, res, -min, "The result should be greater or equal to the minimum")
		assert.LessOrEqual(t, res, max, "The result should be less or equal to the maximum")
	} else {
		// res is negative
		assert.GreaterOrEqual(t, res, -max, "The result should be greater or equal to the minimum")
		assert.LessOrEqual(t, res, min, "The result should be less or equal to the maximum")
	}
}

func Test_GenerateRandomFloatTransformer(t *testing.T) {

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
