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
	minValue := float64(2.2)
	maxValue := float64(4.2)

	val, err := GetFloat64Range(minValue, maxValue)
	assert.NoError(t, err)

	assert.Equal(t, maxValue-minValue, val)
}

func Test_GetFloat64RangeError(t *testing.T) {
	minValue := float64(6.9)
	maxValue := float64(2.2)

	_, err := GetFloat64Range(minValue, maxValue)
	assert.Error(t, err)
}

func Test_GetFloat64RangeMinEqualMax(t *testing.T) {
	minValue := float64(2.2)
	maxValue := float64(2.2)

	val, err := GetFloat64Range(minValue, maxValue)
	assert.NoError(t, err)

	assert.Equal(t, minValue, val)
}

func Test_IsNegativeFloatTrue(t *testing.T) {
	val := IsNegativeFloat64(-1.63)

	assert.True(t, val, "The value should be negative")
}

func Test_IsNegativeFloatFalse(t *testing.T) {
	val := IsNegativeFloat64(324.435)

	assert.False(t, val, "The value should be positive")
}

func Test_GetFloatLength(t *testing.T) {
	val := float64(3.14)
	res := GetFloatLength(val)

	assert.Equal(t, int64(1), GetInt64Length(res.DigitsBeforeDecimalLength), "The actual value should be the same length as the input value")
	assert.Equal(t, int64(1), GetInt64Length(res.DigitsAfterDecimalLength), "The actual value should be the same length as the input value")
}

func Test_ApplyPrecisionToFloat64ErrorTooShort(t *testing.T) {
	precision := 0
	value := float64(2.4356789)

	_, err := ReduceFloat64Precision(precision, value)
	assert.Error(t, err)
}

func Test_ApplyPrecisionToFloat64LengthError(t *testing.T) {
	precision := 20
	value := float64(2.4356789)

	_, err := ReduceFloat64Precision(precision, value)
	assert.Error(t, err)
}

func Test_ApplyPrecisionToFloat64LongPrecision(t *testing.T) {
	precision := 13
	value := float64(2.4356789)

	res, err := ReduceFloat64Precision(precision, value)
	assert.NoError(t, err)

	assert.Equal(t, GetFloat64Length(value), GetFloat64Length(res), "The float should have reduced precision based on the ")
}

func Test_ApplyPrecisionToFloat64(t *testing.T) {
	precision := 4
	value := float64(2.4356789)

	res, err := ReduceFloat64Precision(precision, value)
	assert.NoError(t, err)

	assert.Equal(t, GetFloat64Length(value)-int64(precision), GetFloat64Length(res), "The float should have reduced precision based on the ")
}

func Test_ApplyPrecisionToFloat64Negative(t *testing.T) {
	precision := 4
	value := float64(-2.4356789)

	res, err := ReduceFloat64Precision(precision, value)
	assert.NoError(t, err)

	assert.Equal(t, GetFloat64Length(value)-int64(precision), GetFloat64Length(res), "The float should have reduced precision based on the ")
}
