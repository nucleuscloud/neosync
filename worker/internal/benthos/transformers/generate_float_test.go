package transformers

import (
	"fmt"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateRandomFloatDigitsAfterDecimalErrorTooShort(t *testing.T) {

	_, err := GenerateRandomFloat("positive", 1, 0)

	assert.Error(t, err, "The digits after decimal cannot be less than 1")

}

func Test_GenerateRandomFloatDigitsAfterDecimalErrorTooLong(t *testing.T) {

	_, err := GenerateRandomFloat("positive", 1, 11)

	assert.Error(t, err, "The digits after decimal cannot be greater than 9")

}

func Test_GenerateRandomFloatDigitsBeforeDecimalErrorTooShort(t *testing.T) {

	_, err := GenerateRandomFloat("positive", -1, 4)

	assert.Error(t, err, "The digits after decimal cannot be less than 1")

}

func Test_GenerateRandomFloatDigitsBeforeDecimalErrorTooLong(t *testing.T) {

	_, err := GenerateRandomFloat("positive", 12, 5)

	assert.Error(t, err, "The digits after decimal cannot be greater than 9")

}

func Test_GenerateRandomFloatWrongSign(t *testing.T) {

	_, err := GenerateRandomFloat("nosign", 2, 5)

	assert.Error(t, err, "The sign should be either positive, negative or random")

}
func Test_GenerateRandomFloatPositive(t *testing.T) {

	dbd := 4
	dad := 4

	res, err := GenerateRandomFloat("positive", int64(dbd), int64(dad))
	actual := GetFloatLength(res).DigitsBeforeDecimalLength + GetFloatLength(res).DigitsAfterDecimalLength

	assert.NoError(t, err)
	assert.Equal(t, IsNegativeInt(int64(dbd)), IsNegativeFloat(res), "The expected value should be positive")
	assert.Equal(t, dbd+dad, actual, "The output float needs to be the same length as the input Float")

}

func Test_GenerateRandomFloatRandom(t *testing.T) {

	dbd := 4
	dad := 4

	res, err := GenerateRandomFloat("random", int64(dbd), int64(dad))
	actual := GetFloatLength(res).DigitsBeforeDecimalLength + GetFloatLength(res).DigitsAfterDecimalLength

	assert.NoError(t, err)

	if res < 0 {
		assert.Equal(t, IsNegativeInt(int64(dbd*-1)), IsNegativeFloat(res), "The expected value should be negative")
	} else {
		assert.Equal(t, IsNegativeFloat(float64(dbd+dad)), IsNegativeFloat(res), "The expected value should be positive and 8 digits in length")
	}

	if res < 0 {
		assert.Equal(t, dbd+dad+1, actual, "The output float needs to be the same length as the input Float")

	} else {
		assert.Equal(t, dbd+dad, actual, "The output float needs to be the same length as the input Float")

	}

}

func Test_GenerateRandomFloatNegative(t *testing.T) {

	dbd := 4
	dad := 4

	res, err := GenerateRandomFloat("negative", int64(dbd), int64(dad))

	actual := GetFloatLength(res).DigitsBeforeDecimalLength + GetFloatLength(res).DigitsAfterDecimalLength

	assert.NoError(t, err)
	assert.Equal(t, IsNegativeInt(int64(dbd*-1)), IsNegativeFloat(res), "The expected value should be negative")

	// + 1 to account for the negative signal
	assert.Equal(t, dbd+dad+1, actual, "The output float should be 9 digits long")

}

func Test_GenerateRandomFloatWithLength(t *testing.T) {

	dbd := 4
	dad := 4

	res, err := GenerateRandomFloatWithLength(dbd, dad)

	actual := GetFloatLength(res).DigitsAfterDecimalLength + GetFloatLength(res).DigitsBeforeDecimalLength
	assert.NoError(t, err)
	assert.Equal(t, actual, dbd+dad, "The length of the output float needs to match the digits before + the digits after")
}

func Test_GenerateRandomFloatTransformer(t *testing.T) {

	dbd := 4
	dad := 4
	mapping := fmt.Sprintf(`root = generate_random_float("positive", %d, %d)`, dbd, dad)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the random float transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	actual := GetFloatLength(res.(float64)).DigitsAfterDecimalLength + GetFloatLength(res.(float64)).DigitsBeforeDecimalLength
	assert.Equal(t, dbd+dad, actual, "The length of the output float needs to match the digits before + the digits after")
	assert.Equal(t, IsNegativeInt(int64(dbd)), IsNegativeFloat(res.(float64)), "The expected value should be positive")
	assert.IsType(t, res, float64(1), "The expected value should be a float64")
}

func Test_IsNegativeFloatTrue(t *testing.T) {

	val := IsNegativeFloat(-1.63)

	assert.True(t, val, "The value should be negative")
}

func Test_IsNegativeFloatFalse(t *testing.T) {

	val := IsNegativeFloat(324.435)

	assert.False(t, val, "The value should be positive")
}
