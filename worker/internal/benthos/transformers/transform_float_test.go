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

	actual := GetFloatLength(*res).DigitsAfterDecimalLength + GetFloatLength(*res).DigitsBeforeDecimalLength

	expectedLength := GetFloatLength(testFloatValue).DigitsAfterDecimalLength + GetFloatLength(testFloatValue).DigitsBeforeDecimalLength

	assert.Equal(t, actual, expectedLength, "The length of the output float needs to match the digits before + the digits after")

	assert.Equal(t, IsNegativeFloat(testFloatValue), IsNegativeFloat(*res), "The actual value should be the same sign as the input value")
}

func Test_TransformFloatPreserveLengthFalse(t *testing.T) {

	res, err := TransformFloat(testFloatValue, false, true)
	assert.NoError(t, err)

	expectedLength := GetFloatLength(testFloatValue).DigitsAfterDecimalLength + GetFloatLength(testFloatValue).DigitsBeforeDecimalLength

	assert.Equal(t, defaultDigitsAfterDecimal+defaultDigitsBeforeDecimal, expectedLength, "The length of the output float needs to match the digits before + the digits after")

	assert.Equal(t, IsNegativeFloat(testFloatValue), IsNegativeFloat(*res), "The actual value should be the same sign as the input value")
}

func Test_TransformFloatPreserveSignFalse(t *testing.T) {

	res, err := TransformFloat(testFloatValue, true, false)
	assert.NoError(t, err)

	expectedLength := GetFloatLength(testFloatValue).DigitsAfterDecimalLength + GetFloatLength(testFloatValue).DigitsBeforeDecimalLength

	assert.Equal(t, defaultDigitsAfterDecimal+defaultDigitsBeforeDecimal, expectedLength, "The length of the output float needs to match the digits before + the digits after")

	assert.Equal(t, IsNegativeFloat(testFloatValue), IsNegativeFloat(*res), "The actual value should be positive")
}

func Test_TransformFloatPreserveSignTrueNegative(t *testing.T) {

	res, err := TransformFloat(testNegativeFloatValue, true, true)
	assert.NoError(t, err)

	expectedLength := GetFloatLength(testFloatValue).DigitsAfterDecimalLength + GetFloatLength(testFloatValue).DigitsBeforeDecimalLength

	assert.Equal(t, defaultDigitsAfterDecimal+defaultDigitsBeforeDecimal, expectedLength, "The length of the output float needs to match the digits before + the digits after")

	assert.Equal(t, IsNegativeFloat(testNegativeFloatValue), IsNegativeFloat(*res), "The actual value should be the same sign as the input value")
}

func Test_TransformFloatTransformer(t *testing.T) {
	mapping := fmt.Sprintf(`root = transform_float(value:%f,preserve_length: true,preserve_sign: true)`, testFloatValue)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the random float transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.NotNil(t, res, "The response shouldn't be nil.")

	resFloat, ok := res.(*float64)
	if !ok {
		t.Errorf("Expected *string, got %T", res)
		return
	}

	if resFloat != nil {

		actualLength := GetFloatLength(*resFloat).DigitsAfterDecimalLength + GetFloatLength(*resFloat).DigitsBeforeDecimalLength

		expectedLength := GetFloatLength(testFloatValue).DigitsAfterDecimalLength + GetFloatLength(testFloatValue).DigitsBeforeDecimalLength

		assert.IsType(t, *resFloat, float64(1), "The actual value should be a float64")
		assert.Equal(t, IsNegativeFloat(testFloatValue), IsNegativeFloat(*resFloat), "The float should be positive.")
		assert.Equal(t, expectedLength, actualLength, "The length of the output float needs to match the digits before + the digits after")

	} else {
		t.Error("Pointer is nil, expected a valid float64 pointer")
	}

}

func Test_TransformFloatTransformerWithEmptyValue(t *testing.T) {

	nilFloat := float64(0)
	mapping := fmt.Sprintf(`root = transform_float(value:%f,preserve_length: true,preserve_sign: true)`, nilFloat)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the float transformer")

	_, err = ex.Query(nil)
	assert.NoError(t, err)
}

func Test_GetFloatLength(t *testing.T) {
	val := float64(3.14)
	res := GetFloatLength(val)

	assert.Equal(t, int64(1), transformer_utils.GetIntLength(int64(res.DigitsBeforeDecimalLength)), "The actual value should be the same length as the input value")
	assert.Equal(t, int64(1), transformer_utils.GetIntLength(int64(res.DigitsAfterDecimalLength)), "The actual value should be the same length as the input value")
}
