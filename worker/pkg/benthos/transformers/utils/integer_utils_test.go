package transformer_utils

import (
	"fmt"
	"math"
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
	t.Run("basic range", func(t *testing.T) {
		output, err := GenerateRandomInt64InValueRange(rng.New(time.Now().UnixNano()), 2, 5)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, output, int64(2))
		assert.LessOrEqual(t, output, int64(5))
	})

	t.Run("same min and max", func(t *testing.T) {
		output, err := GenerateRandomInt64InValueRange(rng.New(time.Now().UnixNano()), 2, 2)
		require.NoError(t, err)
		assert.Equal(t, int64(2), output)
	})

	t.Run("negative range", func(t *testing.T) {
		output, err := GenerateRandomInt64InValueRange(rng.New(time.Now().UnixNano()), -9, -2)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, output, int64(-9))
		assert.LessOrEqual(t, output, int64(-2))
	})

	t.Run("crossing zero", func(t *testing.T) {
		output, err := GenerateRandomInt64InValueRange(rng.New(time.Now().UnixNano()), -2, 9)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, output, int64(-2))
		assert.LessOrEqual(t, output, int64(9))
	})

	t.Run("zero range", func(t *testing.T) {
		output, err := GenerateRandomInt64InValueRange(rng.New(time.Now().UnixNano()), 0, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(0), output)
	})

	t.Run("swapped min max", func(t *testing.T) {
		output, err := GenerateRandomInt64InValueRange(rng.New(time.Now().UnixNano()), 5, 2)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, output, int64(2))
		assert.LessOrEqual(t, output, int64(5))
	})

	t.Run("max int64", func(t *testing.T) {
		output, err := GenerateRandomInt64InValueRange(rng.New(time.Now().UnixNano()), int64(0), math.MaxInt64)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, output, int64(0))
		assert.LessOrEqual(t, output, int64(math.MaxInt64))
	})

	t.Run("near max int64", func(t *testing.T) {
		output, err := GenerateRandomInt64InValueRange(rng.New(time.Now().UnixNano()), math.MaxInt64-int64(10), math.MaxInt64)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, output, math.MaxInt64-int64(10))
		assert.LessOrEqual(t, output, int64(math.MaxInt64))
	})

	t.Run("large range near max int64", func(t *testing.T) {
		output, err := GenerateRandomInt64InValueRange(rng.New(time.Now().UnixNano()), math.MaxInt64/int64(2), math.MaxInt64)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, output, math.MaxInt64/int64(2))
		assert.LessOrEqual(t, output, int64(math.MaxInt64))
	})
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
