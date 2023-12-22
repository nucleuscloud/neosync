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
	//nolint:all
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
		return 0, fmt.Errorf("length is too large")
	}

	val, err := GenerateRandomInt64InValueRange(min, max)
	if err != nil {
		return 0, fmt.Errorf("unable to generate a value in the range provided")
	}

	res, err := GenerateRandomInt64FixedLength(val)
	if err != nil {
		return 0, fmt.Errorf("unable to generate a value in the range provided")
	}

	//nolint:all
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

	//nolint:all
	return min + rand.Int63n(rangeVal), nil
}

func FirstDigitIsNine(n int64) bool {
	// Convert the int64 to a string
	str := strconv.FormatInt(n, 10)

	// Check if the string is empty or if the first character is '9'
	if len(str) > 0 && str[0] == '9' {
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

func IsLastInt64DigitZero(n int64) bool {
	// Convert the int64 to a string
	str := strconv.FormatInt(n, 10)

	// Check if the string is empty or if the last character is '0'
	if len(str) > 0 && str[len(str)-1] == '0' {
		return true
	}

	return false
}

func IsNegativeInt64(val int64) bool {
	if (val * -1) < 0 {
		return false
	} else {
		return true
	}
}

func AbsInt64(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}

// Returns the int64 range between the min and max
func GetInt64Range(min, max int64) (int64, error) {

	if min > max {
		return 0, fmt.Errorf("min cannot be greater than max")
	}

	if min == max {
		return min, nil
	}

	return max - min, nil

}

func IsInt64InRandomizationRange(value, rMin, rMax int64) bool {

	if rMin > rMax {
		rMin, rMax = rMax, rMin
	}

	if rMin == rMax {
		return value == rMin
	}

	return value >= rMin && value <= rMax
}
