package transformer_utils

import (
	"fmt"
	"testing"
	"time"

	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GenerateRandomInt64WithFixedLength(t *testing.T) {
	l := int64(5)

	val, err := GenerateRandomInt64FixedLength(rng.New(time.Now().UnixNano()), l)
	assert.NoError(t, err)

	assert.Equal(t, l, GetInt64Length(val), "Actual value to be equal to the input length")
}

func Test_GenerateRandomInt64WithFixedLengthError(t *testing.T) {
	l := int64(29)

	_, err := GenerateRandomInt64FixedLength(rng.New(time.Now().UnixNano()), l)
	assert.Error(t, err, "The int length is greater than 19 and too long")
}

func Test_GenerateRandomInt64InLengthRange(t *testing.T) {
	minValue := int64(3)
	maxValue := int64(7)

	val, err := GenerateRandomInt64InLengthRange(rng.New(time.Now().UnixNano()), minValue, maxValue)
	assert.NoError(t, err)

	assert.GreaterOrEqual(t, GetInt64Length(val), minValue, "The expected value should be greater than or equal to the minimum length.")
	assert.LessOrEqual(t, GetInt64Length(val), maxValue, "The expected value should be less than or equal to the maximum length")
}

func Test_GenerateRandomInt64InLengthRangeError(t *testing.T) {
	minValue := int64(3)
	maxValue := int64(29)
	_, err := GenerateRandomInt64InLengthRange(rng.New(time.Now().UnixNano()), minValue, maxValue)
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
			output, err := GenerateRandomInt64InValueRange(rng.New(time.Now().UnixNano()), tc.min, tc.max)
			assert.NoError(t, err)
			assert.GreaterOrEqual(t, output, tc.min, "%d>=%d was not true. output should be greater than or equal to the min. output: %s", output, tc.min, output)
			assert.LessOrEqual(t, output, tc.max, "%d<=%d was not true. output should be less than or equal to the max. output: %s", output, tc.max, output)
		})
	}
}

func Test_GenerateRandomInt64InValueRange_Swapped_MinMax(t *testing.T) {
	minValue := int64(2)
	maxValue := int64(1)
	output, err := GenerateRandomInt64InValueRange(rng.New(time.Now().UnixNano()), minValue, maxValue)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, output, maxValue)
	assert.LessOrEqual(t, output, minValue)
}

func Test_GetInt64Legth(t *testing.T) {
	expected := 3

	val := GetInt64Length(782)

	assert.Equal(t, int64(expected), val, "The calculated length should match the expected length.")
}

func Test_IsLastIntDigitZero(t *testing.T) {
	type testcase struct {
		input    int
		expected bool
	}
	testcases := []testcase{
		{input: 954670, expected: true},
		{input: 23546789, expected: false},
		{input: 0, expected: true},
		{input: 1, expected: false},
		{input: -1, expected: false},
		{input: -10, expected: true},
	}
	for _, tc := range testcases {
		t.Run(fmt.Sprintf("%d", tc.input), func(t *testing.T) {
			assert.Equal(t, tc.expected, IsLastIntDigitZero(tc.input))
		})
	}
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

func Test_MinInt(t *testing.T) {
	assert.Equal(t, 1, MinInt(1, 2))
	assert.Equal(t, 1, MinInt(2, 1))
	assert.Equal(t, 1, MinInt(1, 1))
	assert.Equal(t, -1, MinInt(-1, 1))
	assert.Equal(t, -1, MinInt(1, -1))
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

func Test_ClampInts(t *testing.T) {
	type testcase struct {
		input    []int
		min      *int
		max      *int
		expected []int
	}

	testcases := []testcase{
		{},
		{[]int{1, 2, 3}, nil, nil, []int{1, 2, 3}},
		{[]int{1, 2, 3}, shared.Ptr(2), shared.Ptr(2), []int{2}},
		{[]int{1, 2, 3, 4, 5}, shared.Ptr(2), shared.Ptr(4), []int{2, 3, 4}},
	}

	for _, tc := range testcases {
		t.Run("", func(t *testing.T) {
			actual := ClampInts(tc.input, tc.min, tc.max)
			require.Equal(t, tc.expected, actual)
		})
	}
}
