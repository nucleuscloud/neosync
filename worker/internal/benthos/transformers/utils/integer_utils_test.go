package transformer_utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GenerateRandomInt64WithFixedLength(t *testing.T) {
	l := int64(5)

	val, err := GenerateRandomInt64FixedLength(l)
	assert.NoError(t, err)

	assert.Equal(t, l, GetInt64Length(val), "Actual value to be equal to the input length")
}

func Test_GenerateRandomInt64WithFixedLengthError(t *testing.T) {
	l := int64(29)

	_, err := GenerateRandomInt64FixedLength(l)
	assert.Error(t, err, "The int length is greater than 19 and too long")
}

func Test_GenerateRandomInt64InLengthRange(t *testing.T) {
	min := int64(3)
	max := int64(7)

	val, err := GenerateRandomInt64InLengthRange(min, max)
	assert.NoError(t, err)

	assert.GreaterOrEqual(t, GetInt64Length(val), min, "The expected value should be greater than or equal to the minimum length.")
	assert.LessOrEqual(t, GetInt64Length(val), max, "The expected value should be less than or equal to the maximum length")
}

func Test_GenerateRandomInt64InLengthRangeError(t *testing.T) {
	min := int64(3)
	max := int64(29)
	_, err := GenerateRandomInt64InLengthRange(min, max)
	assert.Error(t, err, "The int length is greater than 19 and too long")
}

func Test_GenerateRandomInt64InValueRange(t *testing.T) {
	type testcase struct {
		min int64
		max int64
	}
	testcases := []testcase{
		{min: int64(2), max: int64(5)},
		{min: int64(23), max: int64(24)},
		{min: int64(4), max: int64(24)},
		{min: int64(2), max: int64(2)},
		{min: int64(2), max: int64(4)},
		{min: int64(1), max: int64(1)},
		{min: int64(0), max: int64(0)},
		{min: int64(-9), max: int64(-2)},
		{min: int64(-2), max: int64(9)},
	}
	for _, tc := range testcases {
		name := fmt.Sprintf("%s_%d_%d", t.Name(), tc.min, tc.max)
		t.Run(name, func(t *testing.T) {
			output, err := GenerateRandomInt64InValueRange(tc.min, tc.max)
			assert.NoError(t, err)
			assert.GreaterOrEqual(t, output, tc.min, "%d>=%d was not true. output should be greater than or equal to the min. output: %s", output, tc.min, output)
			assert.LessOrEqual(t, output, tc.max, "%d<=%d was not true. output should be less than or equal to the max. output: %s", output, tc.max, output)
		})
	}
}

func Test_GenerateRandomInt64InValueRange_Swapped_MinMax(t *testing.T) {
	min := int64(2)
	max := int64(1)
	output, err := GenerateRandomInt64InValueRange(min, max)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, output, max)
	assert.LessOrEqual(t, output, min)
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

func Test_IsLastInt64DigitZeroTrue(t *testing.T) {
	value := int64(954670)

	res := IsLastIntDigitZero(value)
	assert.Equal(t, res, true, "The last digit is zero.")
}

func Test_IsLastDigitZeroFalse(t *testing.T) {
	value := int64(23546789)

	res := IsLastIntDigitZero(value)
	assert.Equal(t, res, false, "The last digit is not zero.")
}

func Test_GetInt64Range(t *testing.T) {
	min := int64(2)
	max := int64(4)

	val, err := GetIntRange(min, max)
	assert.NoError(t, err)

	assert.Equal(t, max-min, val)
}

func Test_GetInt64RangeError(t *testing.T) {
	min := int64(6)
	max := int64(2)

	_, err := GetIntRange(min, max)
	assert.Error(t, err)
}

func Test_GetInt64RangeMinEqualMax(t *testing.T) {
	min := int64(2)
	max := int64(2)

	val, err := GetIntRange(min, max)
	assert.NoError(t, err)

	assert.Equal(t, min, val)
}

func Test_AbsInt64Positive(t *testing.T) {
	val := int64(7)

	res := AbsInt(val)
	assert.Equal(t, int64(7), res)
}

func Test_AbsInt64Negative(t *testing.T) {
	val := int64(-7)

	res := AbsInt(val)
	assert.Equal(t, int64(7), res)
}

func Test_IsNegativeIntTrue(t *testing.T) {
	val := IsNegativeInt(-1)

	assert.True(t, val, "The value should be negative")
}

func Test_IsNegativeIntFalse(t *testing.T) {
	val := IsNegativeInt(1)

	assert.False(t, val, "The value should be positive")
}

func Test_IsValueInRandomizationRangeTrue(t *testing.T) {
	val := int64(27)
	rMin := int64(22)
	rMax := int64(29)

	res := IsIntInRandomizationRange(val, rMin, rMax)
	assert.Equal(t, true, res, "The value should be in the range")
}

func Test_IsValueInRandomizationRangeFalse(t *testing.T) {
	val := int64(27)
	rMin := int64(22)
	rMax := int64(25)

	res := IsIntInRandomizationRange(val, rMin, rMax)
	assert.Equal(t, false, res, "The value should not be in the range")
}

func Test_MinInt(t *testing.T) {
	assert.Equal(t, 1, MinInt(1, 2))
	assert.Equal(t, 1, MinInt(2, 1))
	assert.Equal(t, 1, MinInt(1, 1))
	assert.Equal(t, -1, MinInt(-1, 1))
	assert.Equal(t, -1, MinInt(1, -1))
}

func Test_MaxInt(t *testing.T) {
	assert.Equal(t, 2, MaxInt(1, 2))
	assert.Equal(t, 2, MaxInt(2, 1))
	assert.Equal(t, 1, MaxInt(1, 1))
	assert.Equal(t, 1, MaxInt(-1, 1))
	assert.Equal(t, 1, MaxInt(1, -1))
}

func Test_Floor(t *testing.T) {
	assert.Equal(t, 3, Floor(2, 3))
	assert.Equal(t, 3, Floor(3, 3))
	assert.Equal(t, 4, Floor(4, 3))
}

func Test_Ceil(t *testing.T) {
	assert.Equal(t, 3, Ceil(3, 4))
	assert.Equal(t, 4, Ceil(4, 4))
	assert.Equal(t, 4, Ceil(5, 4))
}
