package transformer_utils

import (
	"testing"
	"time"

	"github.com/nucleuscloud/neosync/worker/internal/rng"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateRandomFloat64WithInclusiveBoundsMinEqualMax(t *testing.T) {
	v1 := float64(2.2)
	v2 := float64(2.2)

	val, err := GenerateRandomFloat64WithInclusiveBounds(rng.New(time.Now().UnixNano()), v1, v2)
	assert.NoError(t, err, "Did not expect an error when min == max")
	assert.Equal(t, v1, val, "actual value to be equal to min/max")
}

func Test_GenerateRandomFloat64WithInclusiveBoundsPositive(t *testing.T) {
	v1 := float64(2.2)
	v2 := float64(5.2)

	val, err := GenerateRandomFloat64WithInclusiveBounds(rng.New(time.Now().UnixNano()), v1, v2)
	assert.NoError(t, err, "Did not expect an error for valid range")
	assert.True(t, val >= v1 && val <= v2, "actual value to be within the range")
}

func Test_GenerateRandomFloat64WithInclusiveBoundsNegative(t *testing.T) {
	v1 := float64(-2.2)
	v2 := float64(-5.2)

	val, err := GenerateRandomFloat64WithInclusiveBounds(rng.New(time.Now().UnixNano()), v1, v2)

	assert.NoError(t, err, "Did not expect an error for valid range")
	assert.True(t, val <= v1 && val >= v2, "actual value to be within the range")
}

func Test_GenerateRandomFloat64WithBoundsNegativeToPositive(t *testing.T) {
	v1 := float64(-2.3)
	v2 := float64(9.32)

	val, err := GenerateRandomFloat64WithInclusiveBounds(rng.New(time.Now().UnixNano()), v1, v2)

	assert.NoError(t, err, "Did not expect an error for valid range")
	assert.True(t, val >= v1 && val <= v2, "actual value to be within the range")
}
