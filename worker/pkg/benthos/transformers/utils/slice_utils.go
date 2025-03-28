package transformer_utils

import (
	"errors"
	"math"

	"github.com/nucleuscloud/neosync/worker/pkg/rng"
)

const alphanumeric = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890"

/* SLICE MANIPULATION UTILS */

// returns a random index from a one-dimensional slice
func GetRandomValueFromSlice[T any](randomizer rng.Rand, arr []T) (T, error) {
	if len(arr) == 0 {
		var zeroValue T
		return zeroValue, errors.New("slice is empty")
	}

	randomIndex := randomizer.Intn(len(arr))
	return arr[randomIndex], nil
}

// Given two sorted slices and a max value, returns the index from each slice that when summing their values,
// will be equal or as close to the maxValue as possible.
// There could be multiple combinations, but this algorithm attempts to find the closest pair, which finds the
// highest value in each slice that satisfies the maxValue constraint
func FindClosestPair(sortedSlice1, sortedSlice2 []int64, maxValue int64) (leftidx, rightidx int64) {
	// Initialize variables to track the best pair found so far and the best individual value.
	bestPair := [2]int64{-1, -1}        // Initialize to (-1, -1) to indicate failure.
	closestDiff := int64(math.MaxInt64) // Initialize with the largest int64 value.
	maxSum := int64(
		0,
	) // Track the maximum sum less than or equal to maxLength with the smallest difference.

	// Check if any of the lists is empty and handle accordingly
	if len(sortedSlice1) == 0 || len(sortedSlice2) == 0 {
		var nonEmptySlice []int64
		var isSecond bool
		if len(sortedSlice1) == 0 {
			nonEmptySlice = sortedSlice2
			isSecond = true
		} else {
			nonEmptySlice = sortedSlice1
		}
		for idx, val := range nonEmptySlice {
			if val <= maxValue && val > maxSum {
				maxSum = val
				if isSecond {
					bestPair = [2]int64{-1, int64(idx)}
				} else {
					bestPair = [2]int64{int64(idx), -1}
				}
			}
		}
		return bestPair[0], bestPair[1]
	}

	// Iterate through all pairs to find the optimal one.
	for i, val1 := range sortedSlice1 {
		for j, val2 := range sortedSlice2 {
			sum := val1 + val2
			diff := AbsInt(val1 - val2)
			// Check if this pair is within the maxLength and optimizes for closeness.
			if sum <= maxValue && (sum > maxSum || (sum == maxSum && diff < closestDiff)) {
				maxSum = sum
				closestDiff = diff
				bestPair = [2]int64{int64(i), int64(j)}
			}
		}
	}

	return bestPair[0], bestPair[1]
}
