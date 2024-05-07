package transformer_utils

import (
	crypto "crypto/rand"
	"fmt"
	"math"
	"math/big"
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

	minValue := int64(math.Pow10(int(l - 1)))
	maxValue := int64(math.Pow10(int(l))) - 1

	// Generate a random number in the range
	//nolint:gosec
	return minValue + rand.Int63n(maxValue-minValue+1), nil
}

/*
Generates a random int64 with length in the inclusive range of [min, max]. For example, given a length range of [4, 7], possible values will have a length ranging from 4 -> 7 digits.
*/
func GenerateRandomInt64InLengthRange(minValue, maxValue int64) (int64, error) {
	if minValue > maxValue {
		minValue, maxValue = maxValue, minValue
	}

	// Ensure the length doesn't exceed the limit for int64
	if minValue > 19 || maxValue > 19 {
		return 0, fmt.Errorf("length of integer must not exceed 19")
	}

	val, err := GenerateRandomInt64InValueRange(minValue, maxValue)
	if err != nil {
		return 0, fmt.Errorf("unable to generate a value in the range provided [%d:%d]: %w", minValue, maxValue, err)
	}

	res, err := GenerateRandomInt64FixedLength(val)
	if err != nil {
		return 0, fmt.Errorf("unable to generate fixed int64 in the range provided [%d:%d: %w]", minValue, maxValue, err)
	}

	return res, nil
}

/* Generates a random int64 in the inclusive range of [min, max]. For example, given a range of [40, 50], possible values range from 40 -> 50, inclusive. */
func GenerateRandomInt64InValueRange(minValue, maxValue int64) (int64, error) {
	if minValue > maxValue {
		minValue, maxValue = maxValue, minValue
	}

	if minValue == maxValue {
		return minValue, nil
	}

	rangeVal := maxValue - minValue + 1
	//nolint:gosec
	return minValue + rand.Int63n(rangeVal), nil
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

// If n is less than floor, returns floor. Otherwise returns n.
func Floor[T int | int64 | int32 | uint | uint32 | uint64](n, floor T) T {
	if n > floor {
		return n
	}
	return floor
}

// If n is greater than ceiling, returns ceiling. Otherwise returns n.
func Ceil[T int | int64 | int32 | uint | uint32 | uint64 | float32 | float64](n, ceiling T) T {
	if n < ceiling {
		return n
	}
	return ceiling
}

func ClampInts[T int | int32 | int64](input []T, minValue, maxValue *T) []T {
	if minValue == nil && maxValue == nil {
		return input
	}
	filtered := []T{}
	for _, num := range input {
		if minValue != nil && num < *minValue {
			continue
		}

		if maxValue != nil && num > *maxValue {
			continue
		}
		filtered = append(filtered, num)
	}
	return filtered
}

func GenerateCryptoSeed() (int64, error) {
	n, err := crypto.Int(crypto.Reader, big.NewInt(1<<62))
	if err != nil {
		return -1, err
	}
	return n.Int64(), nil
}
