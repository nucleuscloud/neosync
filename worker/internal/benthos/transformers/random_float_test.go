package neosync_transformers

import (
	"fmt"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateRandomFloat(t *testing.T) {

	val := float64(6754.3543)
	expectedLength := 8

	res, err := GenerateRandomFloat(val, true, 3, 4)

	actual := GetFloatLength(res).DigitsBeforeDecimalLength + GetFloatLength(res).DigitsAfterDecimalLength

	assert.NoError(t, err)
	assert.Equal(t, expectedLength, actual, "The output float needs to be the same length as the input Float")

}
func Test_GenerateRandomFloatPreserveLengthFalse(t *testing.T) {

	expectedLength := 6

	res, err := GenerateRandomFloatWithDefinedLength(int64(3), int64(3))

	actual := GetFloatLength(res).DigitsAfterDecimalLength + GetFloatLength(res).DigitsBeforeDecimalLength
	assert.NoError(t, err)
	assert.Equal(t, actual, expectedLength, "The length of the output float needs to match the digits before + the digits after")
}

func Test_GenerateRandomFloatWithRandomLength(t *testing.T) {

	res, err := GenerateRandomFloatWithRandomLength()

	actual := GetFloatLength(res).DigitsAfterDecimalLength + GetFloatLength(res).DigitsBeforeDecimalLength
	assert.NoError(t, err)
	assert.Equal(t, 6, actual, "The length of the output float needs to match the digits before + the digits after")
}

func TestRandomFloatTransformerWithValue(t *testing.T) {
	testVal := float64(397.34)
	mapping := fmt.Sprintf(`root = randomfloattransformer(%f, true, 2,3)`, testVal)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the random float transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.Equal(t, GetFloatLength(testVal), GetFloatLength(res.(float64))) // Generated Float must be the same length as the input Float"
	assert.IsType(t, res, testVal)
}

func TestRandomFloatTransformerWithNoValue(t *testing.T) {
	mapping := `root = randomfloattransformer()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the random float transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	actual := GetFloatLength(res.(float64)).DigitsAfterDecimalLength + GetFloatLength(res.(float64)).DigitsBeforeDecimalLength
	assert.IsType(t, res, float64(1))
	assert.Equal(t, 6, actual, "The length of the output float needs to match the digits before + the digits after")
}

func Test_GetFloatLength(t *testing.T) {
	val := float64(3.14)
	res := GetFloatLength(val)

	assert.Equal(t, int64(1), transformer_utils.GetIntLength(int64(res.DigitsBeforeDecimalLength)))
	assert.Equal(t, int64(1), transformer_utils.GetIntLength(int64(res.DigitsAfterDecimalLength)))

}
