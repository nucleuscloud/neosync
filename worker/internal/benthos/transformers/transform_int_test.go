package transformers

import (
	"fmt"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
	"github.com/stretchr/testify/assert"
)

func Test_TransformIntErrorNotInRange(t *testing.T) {

	val := int64(27)
	rMin := int64(22)
	rMax := int64(25)

	res := transformer_utils.IsInt64InRandomizationRange(val, rMin, rMax)
	assert.Equal(t, false, res, "The value should not be in the range")

}

func Test_TransformIntInRange(t *testing.T) {

	val := int64(27)
	rMin := int64(22)
	rMax := int64(29)

	res, err := TransformInt(val, rMin, rMax)
	assert.NoError(t, err)

	assert.GreaterOrEqual(t, *res, val-rMin, "The result should be greater than the min")
	assert.LessOrEqual(t, *res, val+rMax, "The result should be less than the max")

}

func Test_TransformIntReturnValue(t *testing.T) {

	val := int64(27)
	rMin := int64(27)
	rMax := int64(27)

	res, err := TransformInt(val, rMin, rMax)
	assert.NoError(t, err)

	assert.GreaterOrEqual(t, *res, val-rMin, "The result should be greater than the min")
	assert.LessOrEqual(t, *res, val+rMax, "The result should be less than the max")

}

func Test_TransformIntPhoneTransformerWithNilValue(t *testing.T) {

	val := int64(27)
	rMin := int64(22)
	rMax := int64(29)

	mapping := fmt.Sprintf(`root = transform_int(value:%d, randomization_range_min:%d,randomization_range_max: %d)`, val, rMin, rMax)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the email transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	resInt, ok := res.(*int64)
	if !ok {
		t.Errorf("Expected *int64, got %T", res)
		return
	}

	if resInt != nil {
		assert.GreaterOrEqual(t, *resInt, val-rMin, "The result should be greater than the min")
		assert.LessOrEqual(t, *resInt, val+rMax, "The result should be less than the max")
	} else {
		assert.Error(t, err, "Expected the pointer to resolve to an int64")
	}

}
