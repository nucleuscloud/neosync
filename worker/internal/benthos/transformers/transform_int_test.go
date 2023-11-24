package transformers

import (
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
	"github.com/stretchr/testify/assert"
)

func Test_TransformIntPreserveLengthFalse(t *testing.T) {

	val := int64(67543543)

	res, err := TransformInt(val, false)

	assert.NoError(t, err)
	assert.Equal(t, transformer_utils.GetIntLength(res), int64(4), "The output int needs to be the same length as the input int")

}

func Test_TransformIntError(t *testing.T) {

	val := int64(67567867843543)

	_, err := TransformInt(val, false)

	assert.Error(t, err)

}

func Test_TransformIntPreserveLengthTrue(t *testing.T) {

	val := int64(67543543)

	res, err := TransformInt(val, true)

	assert.NoError(t, err)
	assert.Equal(t, transformer_utils.GetIntLength(res), int64(transformer_utils.GetIntLength((val))), "The output int needs to be the same length as the input int")

}

func Test_TransformIntTransformerWithPreserveLengthFalse(t *testing.T) {
	mapping := `root = transform_int(5, false)`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the random int transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.Equal(t, int64(4), transformer_utils.GetIntLength(res.(int64)), "The expected value shoudl be 4 digits long")
	assert.IsType(t, res, int64(2), "The expected value should be an int64")
}

func Test_TransformIntTransformerWithPreserveLength(t *testing.T) {
	mapping := `root = transform_int(58323, true)`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the random int transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.Equal(t, int64(5), transformer_utils.GetIntLength(res.(int64)), "The expected value shoudl be 5 digits long")
	assert.IsType(t, res, int64(2), "The expected value should be an int64")
}
