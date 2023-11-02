package transformer_utils

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

// returns a random index from a one-dimensional slice
func TestGetRandomValueFromSliceEmptySlice(t *testing.T) {

	arr := []string{}
	_, err := GetRandomValueFromSlice(arr)
	assert.Error(t, err, "Expected an error for the empty slice")

}

func TestGetRandomValueFromSliceNonEmptySlice(t *testing.T) {

	arr := []string{"a", "b", "c"}
	res, err := GetRandomValueFromSlice(arr)
	assert.NoError(t, err)
	assert.Contains(t, arr, res, "Expected the response to be included in the input array")

}

func TestGenerateRandomNumberWithBoundsMinError(t *testing.T) {

	_, err := GenerateRandomIntWithBounds(10, 1)
	assert.Error(t, err, "Expected an error such that the min is greated than the max")
}

func TestGenerateRandomNumberWithBoundsMinEqualMax(t *testing.T) {

	const minMax = 5
	val, err := GenerateRandomIntWithBounds(minMax, minMax)
	assert.NoError(t, err, "Did not expect an error when min == max")
	assert.Equal(t, minMax, val, "Expected value to be equal to min/max")

}

func TestGenerateRandomNumberWithBoundsValid(t *testing.T) {

	min, max := 2, 9
	val, err := GenerateRandomIntWithBounds(min, max)
	assert.NoError(t, err, "Did not expect an error for valid range")
	assert.True(t, val >= min && val <= max, "Expected value to be within the range")
}

func TestSliceStringEmptyString(t *testing.T) {

	res := SliceString("", 10)
	assert.Empty(t, res, "Expected result to be an empty string")
}

func TestSliceStringShortString(t *testing.T) {

	s := "short"
	res := SliceString(s, 10)
	assert.Equal(t, s, res, "Expected result to be equal to the input string")

}

func TestSliceStringValidSlice(t *testing.T) {

	s := "hello, world"
	length := 5
	expected := "hello"
	res := SliceString(s, length)
	assert.Equal(t, expected, res, "Expected result to be a substring of the input string with the specified length")
}

func TestIntArryToStringArr(t *testing.T) {

	val := []int64{1, 2, 3, 4}

	res := IntSliceToStringSlice(val)

	assert.IsType(t, res, []string{})
	assert.Equal(t, len(res), len(val))

}

func TestIntArryToStringArrEmptySlice(t *testing.T) {

	val := []int64{}

	res := IntSliceToStringSlice(val)

	assert.IsType(t, res, []string{})
	assert.Equal(t, len(res), len(val))
}

func TestGenerateRandomInt(t *testing.T) {

	expectedLength := 9

	res, err := GenerateRandomInt(int64(expectedLength))

	assert.NoError(t, err)
	numStr := strconv.FormatInt(res, 10)
	assert.Equal(t, len(numStr), expectedLength, "The length of the generated random int should be the same as the expectedLength")

}

func TestFirstDigitIsNineTrue(t *testing.T) {

	value := int64(9546789)

	res := FirstDigitIsNine(value)
	assert.Equal(t, res, true, "The first digit is nine.")
}

func TestFirstDigitIsNineFalse(t *testing.T) {

	value := int64(23546789)

	res := FirstDigitIsNine(value)
	assert.Equal(t, res, false, "The first digit is not nine.")
}

func TestGetIntLegth(t *testing.T) {

	expected := 3

	val := GetIntLength(782)

	assert.Equal(t, int64(expected), val, "The calculated length should match the expected length.")
}

func TestIsLastDigitZeroTrue(t *testing.T) {

	value := int64(954670)

	res := IsLastDigitZero(value)
	assert.Equal(t, res, true, "The last digit is zero.")
}

func TestIsLastDigitZeroFalse(t *testing.T) {

	value := int64(23546789)

	res := IsLastDigitZero(value)
	assert.Equal(t, res, false, "The last digit is not zero.")
}

func TestRandomStringGeneration(t *testing.T) {

	expectedLength := 5
	res, err := GenerateRandomStringWithLength(int64(expectedLength))

	assert.NoError(t, err)
	assert.Equal(t, len(res), expectedLength, "The output string should be as long as the input string")

}
