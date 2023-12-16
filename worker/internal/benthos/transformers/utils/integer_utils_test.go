package transformer_utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GenerateRandomInt64WithInclusiveBoundsMinEqualMax(t *testing.T) {

	v1 := int64(5)
	v2 := int64(5)
	val, err := GenerateRandomInt64WithInclusiveBounds(v1, v2)
	assert.NoError(t, err, "Did not expect an error when min == max")
	assert.Equal(t, v1, val, "actual value to be equal to min/max")

}

func Test_GenerateRandomInt64WithInclusiveBoundsPositive(t *testing.T) {

	v1 := int64(2)
	v2 := int64(9)

	val, err := GenerateRandomInt64WithInclusiveBounds(v1, v2)
	assert.NoError(t, err, "Did not expect an error for valid range")
	assert.True(t, val >= v1 && val <= v2, "actual value to be within the range")

}

func Test_GenerateRandomInt64WithInclusiveBoundsNegative(t *testing.T) {

	v1 := int64(-2)
	v2 := int64(-9)

	val, err := GenerateRandomInt64WithInclusiveBounds(v1, v2)

	assert.NoError(t, err, "Did not expect an error for valid range")
	assert.True(t, val <= v1 && val >= v2, "actual value to be within the range")

}

func Test_GenerateRandomInt64WithInclusiveBoundsNegativeToPositive(t *testing.T) {

	v1 := int64(-2)
	v2 := int64(9)

	val, err := GenerateRandomInt64WithInclusiveBounds(v1, v2)

	assert.NoError(t, err, "Did not expect an error for valid range")
	assert.True(t, val >= v1 && val <= v2, "actual value to be within the range")
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

func Test_GetInt64Legth(t *testing.T) {

	expected := 3

	val := GetInt64Length(782)

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

func Test_GetInt64Range(t *testing.T) {

	min := int64(2)
	max := int64(4)

	val, err := GetInt64Range(min, max)
	assert.NoError(t, err)

	assert.Equal(t, max-min, val)

}

func Test_GetInt64RangeError(t *testing.T) {

	min := int64(6)
	max := int64(2)

	_, err := GetInt64Range(min, max)
	assert.Error(t, err)

}

func Test_GetInt64RangeMinEqualMax(t *testing.T) {

	min := int64(2)
	max := int64(2)

	val, err := GetInt64Range(min, max)
	assert.NoError(t, err)

	assert.Equal(t, min, val)

}

func Test_AbsInt64Positive(t *testing.T) {

	val := int64(7)

	res := AbsInt64(val)
	assert.Equal(t, int64(7), res)

}

func Test_AbsInt64Negative(t *testing.T) {

	val := int64(-7)

	res := AbsInt64(val)
	assert.Equal(t, int64(7), res)

}

func Test_IsNegativeIntTrue(t *testing.T) {

	val := IsNegativeInt64(-1)

	assert.True(t, val, "The value should be negative")
}

func Test_IsNegativeIntFalse(t *testing.T) {

	val := IsNegativeInt64(1)

	assert.False(t, val, "The value should be positive")
}

func Test_IsValueInRandomizationRangeTrue(t *testing.T) {

	val := int64(27)
	rMin := int64(22)
	rMax := int64(29)

	res := IsInt64InRandomizationRange(val, rMin, rMax)
	assert.Equal(t, true, res, "The value should be in the range")
}

func Test_IsValueInRandomizationRangeFalse(t *testing.T) {

	val := int64(27)
	rMin := int64(22)
	rMax := int64(25)

	res := IsInt64InRandomizationRange(val, rMin, rMax)
	assert.Equal(t, false, res, "The value should not be in the range")

}
