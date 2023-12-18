package transformers

import (
	"fmt"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
	"github.com/stretchr/testify/assert"
)

func Test_TransformFloat64ErrorNotInRange(t *testing.T) {

	val := float64(27.1)
	rMin := float64(22.9)
	rMax := float64(25.333)

	res := transformer_utils.IsFloat64InRandomizationRange(val, rMin, rMax)
	assert.Equal(t, false, res, "The value should not be in the range")

}

func Test_TransformFloat64InRange(t *testing.T) {

	val := float64(27.2323)
	rMin := float64(22.12)
	rMax := float64(29.9823)

	res, err := TransformFloat(val, rMin, rMax)
	assert.NoError(t, err)

	assert.GreaterOrEqual(t, *res, val-rMin, "The result should be greater than the min")
	assert.LessOrEqual(t, *res, val+rMax, "The result should be less than the max")

}

func Test_TransformFloat64PhoneTransformerWithNilValue(t *testing.T) {

	val := float64(27.35)
	rMin := float64(22.24)
	rMax := float64(29.928)

	mapping := fmt.Sprintf(`root = transform_float64(value:%f, randomization_range_min:%f,randomization_range_max: %f)`, val, rMin, rMax)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the email transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	resInt, ok := res.(*float64)
	if !ok {
		t.Errorf("Expected *float64, got %T", res)
		return
	}

	if resInt != nil {
		assert.GreaterOrEqual(t, *resInt, val-rMin, "The result should be greater than the min")
		assert.LessOrEqual(t, *resInt, val+rMax, "The result should be less than the max")
	} else {
		assert.Error(t, err, "Expected the pointer to resolve to an float64")
	}

}
