package transformer_utils

import (
	"fmt"
	"math/rand"
	"strconv"
)

/* INTEGER MANIPULATION UTILS */

// Generates a random int between two integers inclusive of the boundaries
func GenerateRandomInt64WithInclusiveBounds(min, max int64) (int64, error) {

	if min > max {
		min, max = max, min
	}

	if min == max {
		return min, nil
	}

	rangeVal := max - min + 1
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

func IsLastDigitZero(n int64) bool {
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
