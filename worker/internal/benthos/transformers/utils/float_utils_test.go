package transformer_utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GenerateRandomFloat64WithInclusiveBoundsMinEqualMax(t *testing.T) {

	v1 := float64(2.2)
	v2 := float64(2.2)

	val, err := GenerateRandomFloat64WithInclusiveBounds(v1, v2)
	assert.NoError(t, err, "Did not expect an error when min == max")
	assert.Equal(t, v1, val, "actual value to be equal to min/max")

}

func Test_GenerateRandomFloat64WithInclusiveBoundsPositive(t *testing.T) {

	v1 := float64(2.2)
	v2 := float64(5.2)

	val, err := GenerateRandomFloat64WithInclusiveBounds(v1, v2)
	assert.NoError(t, err, "Did not expect an error for valid range")
	assert.True(t, val >= v1 && val <= v2, "actual value to be within the range")

}

func Test_GenerateRandomFloat64WithInclusiveBoundsNegative(t *testing.T) {

	v1 := float64(-2.2)
	v2 := float64(-5.2)

	val, err := GenerateRandomFloat64WithInclusiveBounds(v1, v2)

	assert.NoError(t, err, "Did not expect an error for valid range")
	assert.True(t, val <= v1 && val >= v2, "actual value to be within the range")

}

func Test_GenerateRandomFloat64WithBoundsNegativeToPositive(t *testing.T) {

	v1 := float64(-2.3)
	v2 := float64(9.32)

	val, err := GenerateRandomFloat64WithInclusiveBounds(v1, v2)

	assert.NoError(t, err, "Did not expect an error for valid range")
	assert.True(t, val >= v1 && val <= v2, "actual value to be within the range")

}

func Test_GetFloat64Legth(t *testing.T) {

	expected := 8

	val := GetFloat64Length(7823.2332)

	assert.Equal(t, int64(expected), val, "The calculated length should match the expected length.")
}

func Test_GetFloat64Range(t *testing.T) {

	min := float64(2.2)
	max := float64(4.2)

	val, err := GetFloat64Range(min, max)
	assert.NoError(t, err)

	assert.Equal(t, max-min, val)

}

func Test_GetFloat64RangeError(t *testing.T) {

	min := float64(6.9)
	max := float64(2.2)

	_, err := GetFloat64Range(min, max)
	assert.Error(t, err)

}

func Test_GetFloat64RangeMinEqualMax(t *testing.T) {

	min := float64(2.2)
	max := float64(2.2)

	val, err := GetFloat64Range(min, max)
	assert.NoError(t, err)

	assert.Equal(t, min, val)
}

func Test_IsNegativeFloatTrue(t *testing.T) {

	val := IsNegativeFloat64(-1.63)

	assert.True(t, val, "The value should be negative")
}

func Test_IsNegativeFloatFalse(t *testing.T) {

	val := IsNegativeFloat64(324.435)

	assert.False(t, val, "The value should be positive")
}

func Test_IsFloat64InRandomizationRangeTrue(t *testing.T) {

	val := float64(27)
	rMin := float64(22)
	rMax := float64(29)

	res := IsFloat64InRandomizationRange(val, rMin, rMax)
	assert.Equal(t, true, res, "The value should be in the range")
}

func Test_IsFloat64InRandomizationRangeFalse(t *testing.T) {

	val := float64(27)
	rMin := float64(22)
	rMax := float64(25)

	res := IsFloat64InRandomizationRange(val, rMin, rMax)
	assert.Equal(t, false, res, "The value should not be in the range")

}

func Test_GetFloatLength(t *testing.T) {
	val := float64(3.14)
	res := GetFloatLength(val)

	assert.Equal(t, int64(1), GetInt64Length(int64(res.DigitsBeforeDecimalLength)), "The actual value should be the same length as the input value")
	assert.Equal(t, int64(1), GetInt64Length(int64(res.DigitsAfterDecimalLength)), "The actual value should be the same length as the input value")
}
