package transformer_utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// returns a random index from a one-dimensional slice
func Test_GetRandomValueFromSliceEmptySlice(t *testing.T) {

	arr := []string{}
	_, err := GetRandomValueFromSlice(arr)
	assert.Error(t, err, "Expected an error for the empty slice")

}

func Test_GetRandomValueFromSliceNonEmptySlice(t *testing.T) {

	arr := []string{"a", "b", "c"}
	res, err := GetRandomValueFromSlice(arr)
	assert.NoError(t, err)
	assert.Contains(t, arr, res, "Expected the response to be included in the input array")

}

func Test_IntArryToStringArr(t *testing.T) {

	val := []int64{1, 2, 3, 4}

	res := IntSliceToStringSlice(val)

	assert.IsType(t, res, []string{})
	assert.Equal(t, len(res), len(val), "The slices should be the same length")

}

func Test_IntArryToStringArrEmptySlice(t *testing.T) {

	val := []int64{}

	res := IntSliceToStringSlice(val)

	assert.IsType(t, res, []string{})
	assert.Equal(t, len(res), len(val), "The slices should be the same length")
}
