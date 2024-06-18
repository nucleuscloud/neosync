package transformer_utils

import (
	"testing"
	"time"

	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	"github.com/stretchr/testify/require"
)

func Test_GenerateRandomFloat64WithInclusiveBoundsMinEqualMax(t *testing.T) {
	v1 := float64(2.2)
	v2 := float64(2.2)

	val, err := GenerateRandomFloat64WithInclusiveBounds(rng.New(time.Now().UnixNano()), v1, v2)
	require.NoError(t, err, "Did not expect an error when min == max")
	require.Equal(t, v1, val, "actual value to be equal to min/max")
}

func Test_GenerateRandomFloat64WithInclusiveBoundsPositive(t *testing.T) {
	v1 := float64(2.2)
	v2 := float64(5.2)

	val, err := GenerateRandomFloat64WithInclusiveBounds(rng.New(time.Now().UnixNano()), v1, v2)
	require.NoError(t, err, "Did not expect an error for valid range")
	require.True(t, val >= v1 && val <= v2, "actual value to be within the range")
}

func Test_GenerateRandomFloat64WithInclusiveBoundsNegative(t *testing.T) {
	v1 := float64(-2.2)
	v2 := float64(-5.2)

	val, err := GenerateRandomFloat64WithInclusiveBounds(rng.New(time.Now().UnixNano()), v1, v2)

	require.NoError(t, err, "Did not expect an error for valid range")
	require.True(t, val <= v1 && val >= v2, "actual value to be within the range")
}

func Test_GenerateRandomFloat64WithBoundsNegativeToPositive(t *testing.T) {
	v1 := float64(-2.3)
	v2 := float64(9.32)

	val, err := GenerateRandomFloat64WithInclusiveBounds(rng.New(time.Now().UnixNano()), v1, v2)

	require.NoError(t, err, "Did not expect an error for valid range")
	require.True(t, val >= v1 && val <= v2, "actual value to be within the range")
}

func Test_AnyToFloat64(t *testing.T) {
	inputs := []any{
		"1.0",
		[]byte("1.0"),
		int(1),
		shared.Ptr(int(1)),
		int8(1),
		shared.Ptr(int8(1)),
		int16(1),
		shared.Ptr(int16(1)),
		int32(1),
		shared.Ptr(int32(1)),
		int64(1),
		shared.Ptr(int64(1)),
		uint(1),
		shared.Ptr(uint(1)),
		uint8(1),
		shared.Ptr(uint8(1)),
		uint16(1),
		shared.Ptr(uint16(1)),
		uint32(1),
		shared.Ptr(uint32(1)),
		uint64(1),
		shared.Ptr(uint64(1)),
		float32(1),
		shared.Ptr(float32(1)),
		float64(1),
		shared.Ptr(float64(1)),
		true,
		false,
		shared.Ptr(true),
	}
	for _, input := range inputs {
		output, err := AnyToFloat64(input)
		require.NoError(t, err)
		require.NotNil(t, output)
	}
}
