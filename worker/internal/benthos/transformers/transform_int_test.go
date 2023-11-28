package transformers

import (
	"fmt"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
	"github.com/stretchr/testify/assert"
)

func Test_TransformIntPreserveLengthFalse(t *testing.T) {

	val := int64(67543543)

	res, err := TransformInt(val, false, true)

	assert.NoError(t, err)
	assert.Equal(t, transformer_utils.GetIntLength(*res), int64(4), "The output int needs to be the same length as the input int")

}

func Test_TransformIntError(t *testing.T) {

	val := int64(67567867843543)

	_, err := TransformInt(val, false, true)

	assert.Error(t, err)

}

func Test_TransformIntPreserveLengthTrue(t *testing.T) {

	val := int64(67543543)

	res, err := TransformInt(val, true, true)

	assert.NoError(t, err)
	assert.Equal(t, transformer_utils.GetIntLength(*res), (transformer_utils.GetIntLength((val))), "The output int needs to be the same length as the input int")
	assert.Equal(t, IsNegativeInt(*res), false, "The value return should be positive")

}

func Test_TransformIntPreserveSignTrue(t *testing.T) {

	val := int64(-367)

	res, err := TransformInt(val, true, true)

	assert.NoError(t, err)
	assert.Equal(t, IsNegativeInt(*res), true, "The value return should be negative")

	assert.Equal(t, transformer_utils.GetIntLength(*res), transformer_utils.GetIntLength((val)), "The output int needs to be the same length as the input int")

}

func Test_TransformIntTransformerWithPreserveLengthFalse(t *testing.T) {

	val := 5
	mapping := fmt.Sprintf(`root = transform_int(value:%d, preserve_length:false,preserve_sign: false)`, val)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the random int transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.NotNil(t, res, "The response shouldn't be nil.")

	resInt, ok := res.(*int64)
	if !ok {
		t.Errorf("Expected *string, got %T", res)
		return
	}

	if resInt != nil {

		assert.Equal(t, int64(4), transformer_utils.GetIntLength(*resInt), "The actual value should be 4 digits long")
		assert.IsType(t, *resInt, int64(2), "The actual value should be an int64")

	} else {
		t.Error("Pointer is inl, expected a valid int64 pointer")
	}
}

func Test_TransformIntTransformerWithPreserveLength(t *testing.T) {
	val := 58223
	mapping := fmt.Sprintf(`root = transform_int(value:%d, preserve_length:true,preserve_sign: true)`, val)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the random int transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.NotNil(t, res, "The response shouldn't be nil.")

	resInt, ok := res.(*int64)
	if !ok {
		t.Errorf("Expected *string, got %T", res)
		return
	}

	if resInt != nil {

		assert.Equal(t, int64(5), transformer_utils.GetIntLength(*resInt), "The actual value should be 5 digits long")
		assert.IsType(t, *resInt, int64(2), "The actual value should be an int64")
	} else {
		t.Error("Pointer is nil, expected a valid int64 pointer")
	}
}

func Test_TransformIntPhoneTransformerWithNilValue(t *testing.T) {

	nilNum := 0
	mapping := fmt.Sprintf(`root = transform_int(value:%d, preserve_length:true,preserve_sign: true)`, nilNum)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the email transformer")

	_, err = ex.Query(nil)
	assert.NoError(t, err)
}
