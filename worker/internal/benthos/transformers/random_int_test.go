package transformers

import (
	"fmt"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
	"github.com/stretchr/testify/assert"
)

func TestGenerateRandomIntPreserveLengthTrue(t *testing.T) {

	val := int64(67543543)
	expectedLength := int64(8)

	res, err := GenerateRandomInt(val, true, 0)

	assert.NoError(t, err)
	assert.Equal(t, transformer_utils.GetIntLength(res), expectedLength, "The output int needs to be the same length as the input int")

}

func TestGenerateRandomIntPreserveLengthFalse(t *testing.T) {

	val := int64(67543543)
	expectedLength := int64(4)

	res, err := GenerateRandomInt(val, false, expectedLength)

	assert.NoError(t, err)
	assert.Equal(t, transformer_utils.GetIntLength(res), expectedLength, "The output int needs to be the same length as the input int")

}

func TestGenerateRandomIntPreserveLengthTrueIntLength(t *testing.T) {

	val := int64(67543543)

	_, err := GenerateRandomInt(val, true, int64(5))

	assert.Error(t, err)

}

func TestGenerateandomIntPreserveLengthFalseIntLength(t *testing.T) {

	val := int64(67543543)

	res, err := GenerateRandomInt(val, false, 0)

	assert.NoError(t, err)
	assert.Equal(t, transformer_utils.GetIntLength(res), int64(4), "The output int needs to be the same length as the input int")

}

func TestRandomIntTransformerWithValue(t *testing.T) {
	testVal := int64(397283)
	mapping := fmt.Sprintf(`root = randominttransformer(%d, false, 6)`, testVal)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the random int transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.Equal(t, transformer_utils.GetIntLength(testVal), transformer_utils.GetIntLength(res.(int64))) // Generated int must be the same length as the input int"
	assert.IsType(t, res, testVal)
}

func TestRandomIntTransformerWithNoValue(t *testing.T) {
	mapping := `root = randominttransformer()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the random int transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.Equal(t, int64(4), transformer_utils.GetIntLength(res.(int64))) // Generated int must be the same length as the input int"
	assert.IsType(t, res, int64(2))
}

func TestRandomIntTransformerWithNoValueAndLength(t *testing.T) {
	mapping := `root = randominttransformer(int_length: 4)`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the random int transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.Equal(t, int64(4), transformer_utils.GetIntLength(res.(int64))) // Generated int must be the same length as the input int"
	assert.IsType(t, res, int64(2))
}
