package transformer_utils

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

// returns a random index from a one-dimensional slice
func Test_GetRandomValueFromSliceEmptySlice(t *testing.T) {

	arr := []string{}
	_, err := GetRandomValueFromSlice(arr)
	assert.Error(t, err, "Expected an error for the empty slice")

}

func Test_GetRandomValueFromSliceNonEmptySlice(t *testing.T) {

	arr := []string{"a", "b", "c"}
	res, err := GetRandomValueFromSlice(arr)
	assert.NoError(t, err)
	assert.Contains(t, arr, res, "Expected the response to be included in the input array")

}

func Test_GenerateRandomNumberWithBoundsMinError(t *testing.T) {

	_, err := GenerateRandomIntWithInclusiveBounds(10, 1)
	assert.Error(t, err, "Expected an error such that the min is greated than the max")
}

func Test_GenerateRandomNumberWithBoundsMinEqualMax(t *testing.T) {

	const minMax = 5
	val, err := GenerateRandomIntWithInclusiveBounds(minMax, minMax)
	assert.NoError(t, err, "Did not expect an error when min == max")
	assert.Equal(t, minMax, val, "actual value to be equal to min/max")

}

func Test_GenerateRandomNumberWithBoundsValid(t *testing.T) {

	min, max := 2, 9
	val, err := GenerateRandomIntWithInclusiveBounds(min, max)
	assert.NoError(t, err, "Did not expect an error for valid range")
	assert.True(t, val >= min && val <= max, "actual value to be within the range")
}

func Test_SliceStringEmptyString(t *testing.T) {

	res := SliceString("", 10)
	assert.Empty(t, res, "Expected result to be an empty string")
}

func Test_SliceStringShortString(t *testing.T) {

	s := "short"
	res := SliceString(s, 10)
	assert.Equal(t, s, res, "Expected result to be equal to the input string")

}

func Test_SliceStringValidSlice(t *testing.T) {

	s := "hello, world"
	length := 5
	expected := "hello"
	res := SliceString(s, length)
	assert.Equal(t, expected, res, "Expected result to be a substring of the input string with the specified length")
}

func Test_IntArryToStringArr(t *testing.T) {

	val := []int64{1, 2, 3, 4}

	res := IntSliceToStringSlice(val)

	assert.IsType(t, res, []string{})
	assert.Equal(t, len(res), len(val), "The slices should be the same length")

}

func Test_IntArryToStringArrEmptySlice(t *testing.T) {

	val := []int64{}

	res := IntSliceToStringSlice(val)

	assert.IsType(t, res, []string{})
	assert.Equal(t, len(res), len(val), "The slices should be the same length")
}

func Test_GenerateRandomInt(t *testing.T) {

	expectedLength := 9

	res, err := GenerateRandomInt(expectedLength)

	assert.NoError(t, err)
	numStr := strconv.FormatInt(int64(res), 10)
	assert.Equal(t, len(numStr), expectedLength, "The length of the generated random int should be the same as the expectedLength")

}

func Test_FirstDigitIsNineTrue(t *testing.T) {

	value := int64(9546789)

	res := FirstDigitIsNine(value)
	assert.Equal(t, res, true, "The first digit is nine.")
}

func Test_FirstDigitIsNineFalse(t *testing.T) {

	value := int64(23546789)

	res := FirstDigitIsNine(value)
	assert.Equal(t, res, false, "The first digit is not nine.")
}

func Test_GetIntLegth(t *testing.T) {

	expected := 3

	val := GetIntLength(782)

	assert.Equal(t, int64(expected), val, "The calculated length should match the expected length.")
}

func Test_IsLastDigitZeroTrue(t *testing.T) {

	value := int64(954670)

	res := IsLastDigitZero(value)
	assert.Equal(t, res, true, "The last digit is zero.")
}

func Test_IsLastDigitZeroFalse(t *testing.T) {

	value := int64(23546789)

	res := IsLastDigitZero(value)
	assert.Equal(t, res, false, "The last digit is not zero.")
}

func Test_RandomStringGeneration(t *testing.T) {

	expectedLength := 5
	res, err := GenerateRandomStringWithLength(int64(expectedLength))

	assert.NoError(t, err)
	assert.Equal(t, len(res), expectedLength, "The output string should be as long as the input string")

}
