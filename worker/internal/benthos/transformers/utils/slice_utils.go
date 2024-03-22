package transformer_utils

import (
	"errors"
	"math"
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

func FindClosestPair(sortedSlice1, sortedSlice2 []int64, maxLength int64) (leftidx, rightidx int64) {
	// Initialize variables to track the best pair found so far and the best individual value.
	bestPair := [2]int64{-1, -1}        // Initialize to (-1, -1) to indicate failure.
	closestDiff := int64(math.MaxInt64) // Initialize with the largest int64 value.
	maxSum := int64(0)                  // Track the maximum sum less than or equal to maxLength with the smallest difference.

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
			if val <= maxLength && val > maxSum {
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
			diff := abs(val1 - val2)
			// Check if this pair is within the maxLength and optimizes for closeness.
			if sum <= maxLength && (sum > maxSum || (sum == maxSum && diff < closestDiff)) {
				maxSum = sum
				closestDiff = diff
				bestPair = [2]int64{int64(i), int64(j)}
			}
		}
	}

	return bestPair[0], bestPair[1]
}

// Helper function to calculate the absolute difference
func abs(a int64) int64 {
	if a < 0 {
		return -a
	}
	return a
}

// func FindClosestPair(sortedSlice1, sortedSlice2 []int64, maxLength int64) (leftidx, rightidx int64) {
// 	// Initialize variables to track the best pair found so far.
// 	var bestSum int64 = -1                   // Use -1 to indicate no valid pairs have been found yet
// 	var bestPair [2]int64 = [2]int64{-1, -1} // Initialize to (-1, -1) to indicate failure

// 	// Check for empty slices after attempting to find pairs
// 	if len(sortedSlice1) == 0 && len(sortedSlice2) == 0 {
// 		return -1, -1 // Both slices empty, no pairs possible
// 	}

// 	// Use two pointers to iterate through both slices.
// 	i, j := 0, len(sortedSlice2)-1
// 	for i < len(sortedSlice1) && j >= 0 {
// 		sum := sortedSlice1[i] + sortedSlice2[j]
// 		// Check if this pair is a better solution than what we've found before.
// 		if sum > bestSum && sum <= maxLength {
// 			bestSum = sum
// 			bestPair = [2]int64{int64(i), int64(j)}
// 		}
// 		// Adjust the pointers based on the current sum.
// 		if sum > maxLength {
// 			j--
// 		} else {
// 			i++
// 		}
// 	}

// 	// Only check for individual best elements from each list if no valid pair was found
// 	if bestSum == -1 {
// 		for i, val := range sortedSlice1 {
// 			if val <= maxLength && val > bestSum {
// 				bestSum = val
// 				bestPair = [2]int64{int64(i), -1}
// 			}
// 		}
// 		for j, val := range sortedSlice2 {
// 			if val <= maxLength && val > bestSum {
// 				bestSum = val
// 				bestPair = [2]int64{-1, int64(j)}
// 			}
// 		}
// 	}

// 	// Return the best pair found, preferring pairs over individual elements
// 	return bestPair[0], bestPair[1]
// }

// // Assumes that both slices have already been sorted!!
// func FindClosestPair(sortedSlice1, sortedSlice2 []int64, maxLength int64) (leftidx, rightidx int64) {
// 	// Initialize variables to track the best pair found so far.
// 	var bestSum int64 = -1                   // Use -1 to indicate no valid pairs have been found yet
// 	var bestPair [2]int64 = [2]int64{-1, -1} // Initialize to (-1, -1) to indicate failure

// 	// Check for empty slices
// 	if len(sortedSlice1) == 0 && len(sortedSlice2) == 0 {
// 		return -1, -1 // Both slices empty
// 	} else if len(sortedSlice1) == 0 {
// 		// Only second slice has elements, find if any element is <= maxLength
// 		for j, val := range sortedSlice2 {
// 			if val <= maxLength && val > bestSum {
// 				bestSum = val
// 				bestPair = [2]int64{-1, int64(j)}
// 			}
// 		}
// 		return bestPair[0], bestPair[1]
// 	} else if len(sortedSlice2) == 0 {
// 		// Only first slice has elements, find if any element is <= maxLength
// 		for i, val := range sortedSlice1 {
// 			if val <= maxLength && val > bestSum {
// 				bestSum = val
// 				bestPair = [2]int64{int64(i), -1}
// 			}
// 		}
// 		return bestPair[0], bestPair[1]
// 	}

// 	// Use two pointers to iterate through both slices.
// 	i, j := 0, len(sortedSlice2)-1
// 	for i < len(sortedSlice1) && j >= 0 {
// 		sum := sortedSlice1[i] + sortedSlice2[j]
// 		// Check if this pair is a better solution than what we've found before.
// 		if sum > bestSum && sum <= maxLength {
// 			bestSum = sum
// 			bestPair = [2]int64{int64(i), int64(j)}
// 		}
// 		// Adjust the pointers based on the current sum.
// 		if sum > maxLength {
// 			j--
// 		} else {
// 			i++
// 		}
// 	}

// 	// Check if only elements from one list can make up the bestSum without pairing
// 	for idx, val := range sortedSlice1 {
// 		if val <= maxLength && val > bestSum {
// 			bestSum = val
// 			bestPair = [2]int64{int64(idx), -1}
// 		}
// 	}
// 	for idx, val := range sortedSlice2 {
// 		if val <= maxLength && val > bestSum {
// 			bestSum = val
// 			bestPair = [2]int64{-1, int64(idx)}
// 		}
// 	}

// 	return bestPair[0], bestPair[1]
// }
