package transformers

import (
	"fmt"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateRandomIntPositive(t *testing.T) {
	min := int64(10)
	max := int64(20)

	res, err := GenerateRandomInt64(false, min, max)
	assert.NoError(t, err)

	assert.GreaterOrEqual(t, res, min, "The value should be greater than or equal to the min")
	assert.LessOrEqual(t, res, max, "The value should be less than or equal the max")
}

func Test_GenerateRandomIntNegative(t *testing.T) {
	min := int64(-10)
	max := int64(-20)

	res, err := GenerateRandomInt64(false, min, max)
	assert.NoError(t, err)

	assert.GreaterOrEqual(t, res, max, "The value should be greater than or equal to the min")
	assert.LessOrEqual(t, res, min, "The value should be less than or equal the max")
}

func Test_GenerateRandomIntNegativetoPositive(t *testing.T) {
	min := int64(-10)
	max := int64(20)

	res, err := GenerateRandomInt64(false, min, max)
	assert.NoError(t, err)

	assert.GreaterOrEqual(t, res, min, "The value should be greater than or equal to the min")
	assert.LessOrEqual(t, res, max, "The value should be less than or equal the max")
}

func Test_GenerateRandomIntPositiveRandomSign(t *testing.T) {
	min := int64(10)
	max := int64(20)

	res, err := GenerateRandomInt64(true, min, max)
	assert.NoError(t, err)

	if !transformer_utils.IsNegativeInt64(res) {
		// res is positive
		assert.GreaterOrEqual(t, res, min, "The result should be greater or equal to the minimum")
		assert.LessOrEqual(t, res, max, "The result should be less or equal to the maximum")
	} else {
		// res is negative
		assert.GreaterOrEqual(t, res, -max, "The result should be greater or equal to the minimum")
		assert.LessOrEqual(t, res, -min, "The result should be less or equal to the maximum")
	}
}

func Test_GenerateRandomIntNegativeRandomSign(t *testing.T) {
	min := int64(-10)
	max := int64(-20)

	res, err := GenerateRandomInt64(true, min, max)
	assert.NoError(t, err)

	if !transformer_utils.IsNegativeInt64(res) {
		// res is positive
		assert.GreaterOrEqual(t, res, -min, "The result should be greater or equal to the minimum")
		assert.LessOrEqual(t, res, -max, "The result should be less or equal to the maximum")
	} else {
		// res is negative
		assert.GreaterOrEqual(t, res, max, "The result should be greater or equal to the minimum")
		assert.LessOrEqual(t, res, min, "The result should be less or equal to the maximum")
	}
}

func Test_GenerateRandomIntNegativeToPositiveRandomSign(t *testing.T) {
	min := int64(-10)
	max := int64(20)

	res, err := GenerateRandomInt64(true, min, max)
	assert.NoError(t, err)

	if !transformer_utils.IsNegativeInt64(res) {
		// res is positive
		assert.GreaterOrEqual(t, res, -min, "The result should be greater or equal to the minimum")
		assert.LessOrEqual(t, res, max, "The result should be less or equal to the maximum")
	} else {
		// res is negative
		assert.GreaterOrEqual(t, res, -max, "The result should be greater or equal to the minimum")
		assert.LessOrEqual(t, res, min, "The result should be less or equal to the maximum")
	}
}

func Test_GenerateRandomIntRandomSign(t *testing.T) {
	min := int64(2)
	max := int64(9)
	randomizeSign := false

	mapping := fmt.Sprintf(`root = generate_int64(randomize_sign:%t, min:%d, max:%d)`, randomizeSign, min, max)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the random int transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.GreaterOrEqual(t, res, min, "The result should be greater or equal to the minimum")
	assert.LessOrEqual(t, res, max, "The result should be less or equal to the maximum")
}
