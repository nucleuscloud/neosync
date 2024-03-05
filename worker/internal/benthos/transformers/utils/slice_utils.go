package transformer_utils

import (
	"errors"
	"math/rand"
	"strconv"
)

const alphanumeric = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890"

/* SLICE MANIPULATION UTILS */

// returns a random index from a one-dimensional slice
func GetRandomValueFromSlice[T any](arr []T) (T, error) {
	if len(arr) == 0 {
		var zeroValue T
		return zeroValue, errors.New("slice is empty")
	}

	//nolint:gosec
	randomIndex := rand.Intn(len(arr))

	return arr[randomIndex], nil
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
