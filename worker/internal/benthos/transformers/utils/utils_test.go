package transformer_utils

import (
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

	_, err := GenerateRandomNumberWithBounds(10, 1)
	assert.Error(t, err, "Expected an error such that the min is greated than the max")
}

func TestGenerateRandomNumberWithBoundsMinEqualMax(t *testing.T) {

	const minMax = 5
	val, err := GenerateRandomNumberWithBounds(minMax, minMax)
	assert.NoError(t, err, "Did not expect an error when min == max")
	assert.Equal(t, minMax, val, "Expected value to be equal to min/max")

}

func TestGenerateRandomNumberWithBoundsValid(t *testing.T) {

	min, max := 2, 9
	val, err := GenerateRandomNumberWithBounds(min, max)
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
