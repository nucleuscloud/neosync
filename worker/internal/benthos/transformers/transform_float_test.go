package transformers

import (
	"fmt"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
	"github.com/stretchr/testify/assert"
)

var testFloatValue = float64(3524.24)
var testNegativeFloatValue = float64(-3524.24)

func Test_TransformFloatPreserveLengthTrue(t *testing.T) {

	res, err := TransformFloat(testFloatValue, true, true)
	assert.NoError(t, err)

	actual := GetFloatLength(res).DigitsAfterDecimalLength + GetFloatLength(res).DigitsBeforeDecimalLength

	expectedLength := GetFloatLength(testFloatValue).DigitsAfterDecimalLength + GetFloatLength(testFloatValue).DigitsBeforeDecimalLength

	assert.Equal(t, actual, expectedLength, "The length of the output float needs to match the digits before + the digits after")

	assert.Equal(t, IsNegativeFloat(testFloatValue), IsNegativeFloat(res), "The expected value should be the same sign as the input value")
}

func Test_TransformFloatPreserveLengthFalse(t *testing.T) {

	res, err := TransformFloat(testFloatValue, false, true)
	assert.NoError(t, err)

	expectedLength := GetFloatLength(testFloatValue).DigitsAfterDecimalLength + GetFloatLength(testFloatValue).DigitsBeforeDecimalLength

	assert.Equal(t, defaultDigitsAfterDecimal+defaultDigitsBeforeDecimal, expectedLength, "The length of the output float needs to match the digits before + the digits after")

	assert.Equal(t, IsNegativeFloat(testFloatValue), IsNegativeFloat(res), "The expected value should be the same sign as the input value")
}

func Test_TransformFloatPreserveSignFalse(t *testing.T) {

	res, err := TransformFloat(testFloatValue, true, false)
	assert.NoError(t, err)

	expectedLength := GetFloatLength(testFloatValue).DigitsAfterDecimalLength + GetFloatLength(testFloatValue).DigitsBeforeDecimalLength

	assert.Equal(t, defaultDigitsAfterDecimal+defaultDigitsBeforeDecimal, expectedLength, "The length of the output float needs to match the digits before + the digits after")

	assert.Equal(t, IsNegativeFloat(testFloatValue), IsNegativeFloat(res), "The expected value should be positive")
}

func Test_TransformFloatPreserveSignTrueNegative(t *testing.T) {

	res, err := TransformFloat(testNegativeFloatValue, true, true)
	assert.NoError(t, err)

	expectedLength := GetFloatLength(testFloatValue).DigitsAfterDecimalLength + GetFloatLength(testFloatValue).DigitsBeforeDecimalLength

	assert.Equal(t, defaultDigitsAfterDecimal+defaultDigitsBeforeDecimal, expectedLength, "The length of the output float needs to match the digits before + the digits after")

	assert.Equal(t, IsNegativeFloat(testNegativeFloatValue), IsNegativeFloat(res), "The expected value should be the same sign as the input value")
}

func Test_TransformFloatTransformer(t *testing.T) {
	mapping := fmt.Sprintf(`root = transform_float(%f, true, true)`, testFloatValue)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the random float transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	actualLength := GetFloatLength(res.(float64)).DigitsAfterDecimalLength + GetFloatLength(res.(float64)).DigitsBeforeDecimalLength

	expectedLength := GetFloatLength(testFloatValue).DigitsAfterDecimalLength + GetFloatLength(testFloatValue).DigitsBeforeDecimalLength

	assert.IsType(t, res, float64(1), "The expected value should be a float64")
	assert.Equal(t, IsNegativeFloat(testFloatValue), IsNegativeFloat(res.(float64)))
	assert.Equal(t, expectedLength, actualLength, "The length of the output float needs to match the digits before + the digits after")
}

func Test_GetFloatLength(t *testing.T) {
	val := float64(3.14)
	res := GetFloatLength(val)

	assert.Equal(t, int64(1), transformer_utils.GetIntLength(int64(res.DigitsBeforeDecimalLength)))
	assert.Equal(t, int64(1), transformer_utils.GetIntLength(int64(res.DigitsAfterDecimalLength)))

}
