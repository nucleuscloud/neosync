package transformers

import (
	"fmt"
	"testing"
	"time"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/nucleuscloud/neosync/worker/internal/rng"
	"github.com/stretchr/testify/require"
)

func Test_TransformFloat64InRange(t *testing.T) {
	val := float64(27.2323)
	rMin := float64(5)
	rMax := float64(5)

	res, err := transformFloat(rng.New(time.Now().UnixNano()), &val, rMin, rMax, nil, nil)
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
