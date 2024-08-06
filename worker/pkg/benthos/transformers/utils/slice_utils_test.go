package transformer_utils

import (
	"testing"
	"time"

	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// returns a random index from a one-dimensional slice
func Test_GetRandomValueFromSliceEmptySlice(t *testing.T) {
	arr := []string{}
	_, err := GetRandomValueFromSlice(rng.New(time.Now().UnixNano()), arr)
	assert.Error(t, err, "Expected an error for the empty slice")
}

func Test_GetRandomValueFromSliceNonEmptySlice(t *testing.T) {
	arr := []string{"a", "b", "c"}
	res, err := GetRandomValueFromSlice(rng.New(time.Now().UnixNano()), arr)
	assert.NoError(t, err)
	assert.Contains(t, arr, res, "Expected the response to be included in the input array")
}

func Test_FindClosestPair(t *testing.T) {
	type testcase struct {
		slice1    []int64
		slice2    []int64
		maxLength int64

		expectedLeft  int64
		expectedRight int64
	}

	testcases := []testcase{
		{slice1: []int64{}, slice2: []int64{}, maxLength: 5, expectedLeft: -1, expectedRight: -1},
		{slice1: []int64{5}, slice2: []int64{}, maxLength: 5, expectedLeft: 0, expectedRight: -1},
		{slice1: []int64{}, slice2: []int64{5}, maxLength: 5, expectedLeft: -1, expectedRight: 0},
		{slice1: []int64{1, 2, 3, 4}, slice2: []int64{4}, maxLength: 5, expectedLeft: 0, expectedRight: 0},
		{slice1: []int64{1, 2, 3, 4}, slice2: []int64{3}, maxLength: 4, expectedLeft: 0, expectedRight: 0},
		{slice1: []int64{1, 2, 3, 4}, slice2: []int64{1, 2, 3, 4}, maxLength: 4, expectedLeft: 1, expectedRight: 1},
		{slice1: []int64{1, 2, 3, 4, 5}, slice2: []int64{1, 2, 3, 4, 5}, maxLength: 5, expectedLeft: 1, expectedRight: 2},
		{slice1: []int64{5}, slice2: []int64{5}, maxLength: 4, expectedLeft: -1, expectedRight: -1},
	}

	for _, tc := range testcases {
		t.Run("", func(t *testing.T) {
			actualLeft, actualRight := FindClosestPair(tc.slice1, tc.slice2, tc.maxLength)
			require.Equal(t, tc.expectedLeft, actualLeft)
			require.Equal(t, tc.expectedRight, actualRight)
			var leftSum int64
			var rightSum int64
			if tc.expectedLeft > -1 {
				leftSum = tc.slice1[tc.expectedLeft]
			}
			if tc.expectedRight > -1 {
				rightSum = tc.slice2[tc.expectedRight]
			}
			if tc.expectedLeft > -1 || tc.expectedRight > -1 {
				require.Equal(t, tc.maxLength, leftSum+rightSum)
			}
		})
	}
}
