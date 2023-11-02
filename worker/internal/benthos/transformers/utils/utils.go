package transformer_utils

import (
	"crypto/rand"
	"errors"
	"math/big"
)

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

func GenerateRandomNumberWithBounds(min, max int) (int, error) {

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

func SliceString(s string, l int) string {

	// use runes instead of strings in order to avoid slicing a multi-byte character and returning invalid UTF-8
	runes := []rune(s)

	if l > len(runes) {
		l = len(runes)
	}

	return string(runes[:l])
}
