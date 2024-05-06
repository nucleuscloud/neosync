package transformer_utils

import (
	"github.com/nucleuscloud/neosync/worker/internal/rng"
)

/* FLOAT MANIPULATION UTILS */

// Generates a random float64 in the range of the min and max float64 values
func GenerateRandomFloat64WithInclusiveBounds(randomizer rng.Rand, minValue, maxValue float64) (float64, error) {
	if minValue > maxValue {
		minValue, maxValue = maxValue, minValue
	}

	if minValue == maxValue {
		return minValue, nil
	}

	randValue := randomizer.Float64()

	// Scale and shift the value to the range
	returnValue := minValue + randValue*(maxValue-minValue)
	return returnValue, nil
}
