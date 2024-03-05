package transformer_utils

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
)

/* Generates a random int64 of length l. For example, given a length of 4, possible values will always have a length of 4 digits. */
func GenerateRandomInt64FixedLength(l int64) (int64, error) {
	if l <= 0 {
		return 0, fmt.Errorf("the length has to be greater than zero")
	}

	// Ensure the length doesn't exceed the limit for int64
	if l > 19 {
		return 0, fmt.Errorf("length is too large")
	}

	min := int64(math.Pow10(int(l - 1)))
	max := int64(math.Pow10(int(l))) - 1

	// Generate a random number in the range
	//nolint:gosec
	return min + rand.Int63n(max-min+1), nil
}

/*
Generates a random int64 with length in the inclusive range of [min, max]. For example, given a length range of [4, 7], possible values will have a length ranging from 4 -> 7 digits.
*/
func GenerateRandomInt64InLengthRange(min, max int64) (int64, error) {
	if min > max {
		min, max = max, min
	}

	// Ensure the length doesn't exceed the limit for int64
	if min > 19 || max > 19 {
		return 0, fmt.Errorf("length of integer must not exceed 19")
	}

	val, err := GenerateRandomInt64InValueRange(min, max)
	if err != nil {
		return 0, fmt.Errorf("unable to generate a value in the range provided [%d:%d]: %w", min, max, err)
	}

	res, err := GenerateRandomInt64FixedLength(val)
	if err != nil {
		return 0, fmt.Errorf("unable to generate fixed int64 in the range provided [%d:%d: %w]", min, max, err)
	}

	return res, nil
}

/* Generates a random int64 in the inclusive range of [min, max]. For example, given a range of [40, 50], possible values range from 40 -> 50, inclusive. */
func GenerateRandomInt64InValueRange(min, max int64) (int64, error) {
	if min > max {
		min, max = max, min
	}

	if min == max {
		return min, nil
	}

	rangeVal := max - min + 1
	//nolint:gosec
	return min + rand.Int63n(rangeVal), nil
}

func FirstDigitIsNine(n int64) bool {
	// Convert the int64 to a string
	str := strconv.FormatInt(n, 10)

	// Check if the string is empty or if the first character is '9'
	if str != "" && str[0] == '9' {
		return true
	}

	return false
}

// gets the number of digits in an int64
func GetInt64Length(i int64) int64 {
	// Convert the int64 to a string
	str := strconv.FormatInt(i, 10)
	length := int64(len(str))
	return length
}

func IsLastIntDigitZero[T int | int64 | int32 | uint | uint32 | uint64](n T) bool {
	return n%10 == 0
}

func IsNegativeInt[T int | int64 | int32 | uint | uint32 | uint64](val T) bool {
	return val < 0
}

func AbsInt[T int | int64 | int32 | uint | uint32 | uint64](n T) T {
	if n < 0 {
		return -n
	}
	return n
}

// Returns the int of type T range between the min and max
func GetIntRange[T int | int64 | int32 | uint | uint32 | uint64](min, max T) (T, error) {
	if min > max {
		return 0, fmt.Errorf("min cannot be greater than max")
	}

	if min == max {
		return min, nil
	}

	return max - min, nil
}

func IsIntInRandomizationRange[T int | int64 | int32 | uint | uint32 | uint64](value, rMin, rMax T) bool {
	if rMin > rMax {
		rMin, rMax = rMax, rMin
	}

	if rMin == rMax {
		return value == rMin
	}

	return value >= rMin && value <= rMax
}

func MinInt[T int | int64 | int32 | uint | uint32 | uint64](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func MaxInt[T int | int64 | int32 | uint | uint32 | uint64](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func Floor[T int | int64 | int32 | uint | uint32 | uint64](n, floor T) T {
	if n > floor {
		return n
	}
	return floor
}

func Ceil[T int | int64 | int32 | uint | uint32 | uint64](n, ceiling T) T {
	if n < ceiling {
		return n
	}
	return ceiling
}
