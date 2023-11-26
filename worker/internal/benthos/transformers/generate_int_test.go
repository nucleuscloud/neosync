package transformers

import (
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateRandomIntDefualtLength(t *testing.T) {

	val := int64(67543543)

	_, err := GenerateRandomInt(val, "positive")

	assert.Error(t, err, "The sign should be either positive, negative or random")

}

func Test_GenerateRandomIntWrongSign(t *testing.T) {

	val := int64(67543543)

	_, err := GenerateRandomInt(val, "nosign")

	assert.Error(t, err, "The sign should be either positive, negative or random")

}

func Test_GenerateRandomIntNegativeLength(t *testing.T) {

	val := int64(-1)

	_, err := GenerateRandomInt(val, "positive")

	assert.Error(t, err, "The integer length can't be less than 1")

}

func Test_GenerateRandomIntLengthTooLong(t *testing.T) {

	val := int64(5678976578965789)

	_, err := GenerateRandomInt(val, "positive")

	assert.Error(t, err, "The int length cannot be greater than 18")

}

func Test_GenerateRandomIntRandomSign(t *testing.T) {
	mapping := `root = generate_int(6, "random")`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the random int transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	if res.(int64) < 0 {
		assert.Equal(t, int64(6), transformer_utils.GetIntLength(res.(int64)*-1), "The actual value should be negative and 6 digits in length")
	} else {
		assert.Equal(t, int64(6), transformer_utils.GetIntLength(res.(int64)), "The actual value should be positive and 6 digits in length")
	}

	assert.IsType(t, res, int64(2), "The actual value should be an int64")
}

func Test_GenerateRandomIntTransformerWithNoLength(t *testing.T) {

	mapping := `root = generate_int(6,"negative")`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the random int transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.Equal(t, int64(6), transformer_utils.GetIntLength(res.(int64)*-1))
	assert.Equal(t, IsNegativeInt(res.(int64)), true, "The value return should be negative")
	assert.IsType(t, res, int64(2), "The actual value type should be an int64")
}

func Test_GenerateRandomIntTransformerWithLength(t *testing.T) {
	mapping := `root = generate_int(5, "positive")`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the random int transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.Equal(t, int64(5), transformer_utils.GetIntLength(res.(int64)), "The value should be the same length as the intLength")
	assert.Equal(t, IsNegativeInt(res.(int64)), false, "The value return should be positive")
	assert.IsType(t, res, int64(2), "The value should be an int64")
}

func Test_IsNegativeIntTrue(t *testing.T) {

	val := IsNegativeInt(-1)

	assert.True(t, val, "The value should be negative")
}

func Test_IsNegativeIntFalse(t *testing.T) {

	val := IsNegativeInt(1)

	assert.False(t, val, "The value should be positive")
}
