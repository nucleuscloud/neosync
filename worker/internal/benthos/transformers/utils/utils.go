package transformer_utils

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strconv"
)

const alphanumeric = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890"

// returns a random index from a one-dimensional slice
func GetRandomValueFromSlice[T any](arr []T) (T, error) {
	if len(arr) == 0 {
		var zeroValue T
		return zeroValue, errors.New("slice is empty")
	}

	randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(arr))))
	if err != nil {
		var zeroValue T
		return zeroValue, err
	}

	return arr[randomIndex.Int64()], nil
}

// generates a random int between two numbers
func GenerateRandomIntWithBounds(min, max int) (int, error) {

	min64 := int64(min)
	max64 := int64(max)

	if min > max {
		return 0, errors.New("min cannot be greater than max")
	}

	if min == max {
		return min, nil
	}

	// Generate a random number in the range [0, max-min]
	num, err := rand.Int(rand.Reader, big.NewInt(max64-min64+1))
	if err != nil {
		return 0, err
	}

	// Shift the range to [min, max]
	return int(num.Int64() + min64), nil

}

// substrings a string using rune length to account for multi-byte characters
func SliceString(s string, l int) string {

	// use runes instead of strings in order to avoid slicing a multi-byte character and returning invalid UTF-8
	runes := []rune(s)

	if l > len(runes) {
		l = len(runes)
	}

	return string(runes[:l])
}

// converts a slice of int to a slice of strings
func IntSliceToStringSlice(ints []int64) []string {

	var str []string

	if len(ints) == 0 {
		return []string{}
	}

	for i := range ints {
		str = append(str, strconv.Itoa((i)))

	}

	return str
}

// generates a random integer of length l that is passed in as a int64 param i.e. an l of 3 will generate
// an int64 of 3 digits such as 123 or 789.
func GenerateRandomInt(l int64) (int64, error) {
	if l <= 0 {
		return 0, errors.New("count is zero or not a positive integer")
	}

	// int64 only supports 18 digits, so if the count => 19, this will error out
	if l >= 19 {
		return 0, errors.New("count has to be less than 18 digits since int64 only supports up to 18 digits")
	}

	// Calculate the min and max values for count
	minValue := new(big.Int).Exp(big.NewInt(10), big.NewInt(l-1), nil)
	maxValue := new(big.Int).Exp(big.NewInt(10), big.NewInt(l), nil)

	// Generate a random integer within the specified range
	randInt, err := rand.Int(rand.Reader, maxValue)
	if err != nil {
		return 0, errors.New("unable to generate a random integer")
	}

	/*
		rand.Int generates a random number within the range [0, max-1], so if count == 8 [0 -> 9999999]. If the generated random integer is already the maximum possible value, then adding the minimum value to it will overflow it to count + 1. This is because the big.Int.Add() function adds two big integers together and returns a new big integer. If the first digit is a 9 and it's already count long then adding the min will overflow. So we only add if the digit count is not digits AND the first digit is not 9.

	*/

	if FirstDigitIsNine(randInt.Int64()) && GetIntLength(randInt.Int64()) == l {
		return randInt.Int64(), nil
	} else {
		randInt.Add(randInt, minValue)
		return randInt.Int64(), nil

	}
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
func GetIntLength(i int64) int64 {
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

// generate a random string of length l
func GenerateRandomStringWithLength(l int64) (string, error) {

	if l <= 0 {
		return "", fmt.Errorf("the length cannot be zero or negative")
	}

	// Create a random source using crypto/rand
	source := rand.Reader

	// Calculate the max index in the alphabet string
	maxIndex := big.NewInt(int64(len(alphanumeric)))

	result := make([]byte, l)

	for i := int64(0); i < l; i++ {
		// Generate a random index in the range [0, len(alphabet))
		index, err := rand.Int(source, maxIndex)
		if err != nil {
			return "", fmt.Errorf("unable to generate a random index for random string generation")
		}

		// Get the character at the generated index and append it to the result
		result[i] = alphanumeric[index.Int64()]
	}

	return string(result), nil

}
