package transformers

import (
	"fmt"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/require"
)

func Test_TransformIntInRange(t *testing.T) {
	val := int64(27)
	rMin := int64(5)
	rMax := int64(5)

	res, err := transformInt(&val, rMin, rMax)
	require.NoError(t, err)

	require.GreaterOrEqual(t, *res, val-rMin, "The result should be greater than the min")
	require.LessOrEqual(t, *res, val+rMax, "The result should be less than the max")
}

func Test_TransformInt64_Benthos(t *testing.T) {
	val := int64(27)
	rMin := int64(5)
	rMax := int64(5)

	mapping := fmt.Sprintf(`root = transform_int64(value:%d, randomization_range_min:%d,randomization_range_max: %d)`, val, rMin, rMax)
	ex, err := bloblang.Parse(mapping)
	require.NoError(t, err, "failed to parse the email transformer")

	res, err := ex.Query(nil)
	require.NoError(t, err)

	resInt, ok := res.(*int64)
	if !ok {
		t.Errorf("Expected *int64, got %T", res)
		return
	}

	if resInt != nil {
		require.GreaterOrEqual(t, *resInt, val-rMin, "The result should be greater than the min")
		require.LessOrEqual(t, *resInt, val+rMax, "The result should be less than the max")
	} else {
		require.Error(t, err, "Expected the pointer to resolve to an int64")
	}
}
