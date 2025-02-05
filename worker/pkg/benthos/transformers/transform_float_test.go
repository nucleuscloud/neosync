package transformers

import (
	"fmt"
	"testing"
	"time"

	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/require"
)

func Test_TransformFloat64InRange(t *testing.T) {
	val := float64(27.2323)
	rMin := float64(5)
	rMax := float64(5)

	res, err := transformFloat(rng.New(time.Now().UnixNano()), newMaxNumCache(), &val, rMin, rMax, nil, nil)
	require.NoError(t, err)

	require.GreaterOrEqual(t, *res, val-rMin, "The result should be greater than the min")
	require.LessOrEqual(t, *res, val+rMax, "The result should be less than the max")
}

func Test_TransformFloat64_Benthos(t *testing.T) {
	val := float64(27.35)
	rMin := float64(22.24)
	rMax := float64(29.928)

	mapping := fmt.Sprintf(`root = transform_float64(value:%f, randomization_range_min:%f,randomization_range_max: %f)`, val, rMin, rMax)
	ex, err := bloblang.Parse(mapping)
	require.NoError(t, err, "failed to parse the email transformer")

	res, err := ex.Query(nil)
	require.NoError(t, err)

	resInt, ok := res.(*float64)
	if !ok {
		t.Errorf("Expected *float64, got %T", res)
		return
	}

	if resInt != nil {
		require.GreaterOrEqual(t, *resInt, val-rMin, "The result should be greater than the min")
		require.LessOrEqual(t, *resInt, val+rMax, "The result should be less than the max")
	} else {
		require.Error(t, err, "Expected the pointer to resolve to an float64")
	}
}

func Test_TransformFloat64_Benthos_NoOptions(t *testing.T) {
	val := float64(27.35)
	mapping := fmt.Sprintf(`root = transform_float64(value:%f)`, val)
	ex, err := bloblang.Parse(mapping)
	require.NoError(t, err, "failed to parse the email transformer")

	res, err := ex.Query(nil)
	require.NoError(t, err)
	require.NotEmpty(t, res)
}

func Test_calculateMaxNumber(t *testing.T) {
	tests := []struct {
		precision int
		scale     *int
		expected  float64
		expectErr bool
	}{
		// Valid cases
		{precision: 5, scale: nil, expected: 99999, expectErr: false},                  // Precision 5, scale nil (defaults to 0)
		{precision: 5, scale: shared.Ptr(0), expected: 99999, expectErr: false},        // Precision 5, scale 0
		{precision: 5, scale: shared.Ptr(2), expected: 999.99, expectErr: false},       // Precision 5, scale 2
		{precision: 10, scale: shared.Ptr(3), expected: 9999999.999, expectErr: false}, // Precision 10, scale 3
		{precision: 5, scale: shared.Ptr(-1), expected: 99999, expectErr: false},       // Precision 5, scale 0

		// Invalid cases
		{precision: 0, scale: nil, expected: 0, expectErr: true},           // Invalid precision
		{precision: 3, scale: shared.Ptr(5), expected: 0, expectErr: true}, // Scale greater than precision
	}

	// Run each test
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result, err := calculateMaxNumber(tt.precision, tt.scale)

			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.expected, result)
		})
	}
}
