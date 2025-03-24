package transformers

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

var name = "evis"
var maxCharacterLimit = int64(20)

func Test_TranformFirstNameEmptyName(t *testing.T) {
	randomizer := rng.New(1)
	emptyName := ""

	res, err := transformFirstName(randomizer, emptyName, false, maxCharacterLimit)
	assert.NoError(t, err)
	assert.Empty(t, res, "The response should be empty")
}

func Test_TranformFirstName_Random(t *testing.T) {
	randomizer := rng.New(time.Now().UnixNano())

	res, err := transformFirstName(randomizer, "foo", false, maxCharacterLimit)
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
}

func Test_TransformFirstName_Preserve_True(t *testing.T) {
	randomizer := rng.New(1)

	nameLength := int64(len(name))

	res, err := transformFirstName(randomizer, name, true, maxCharacterLimit)

	assert.NoError(t, err)
	assert.Equal(t, nameLength, int64(len(*res)), "The first name output should be the same length as the input")
	assert.IsType(t, "", *res, "The first name should be a string")
}

func Test_TransformFirstName_Preserve_True_With_Padding(t *testing.T) {
	randomizer := rng.New(1)
	length := 300

	var buffer bytes.Buffer
	char := "a"
	for i := 0; i < length; i++ {
		buffer.WriteString(char)
	}

	name := buffer.String()

	res, err := transformFirstName(randomizer, name, true, 500)

	assert.NoError(t, err)
	assert.Equal(t, int64(len(name)), int64(len(*res)), "The first name output should be the same length as the input")
	assert.IsType(t, "", *res, "The first name should be a string")
}

func Test_TransformFirstName_Preserve_False(t *testing.T) {
	randomizer := rng.New(1)

	res, err := transformFirstName(randomizer, name, false, maxCharacterLimit)

	assert.NoError(t, err)
	assert.LessOrEqual(t, int64(len(*res)), maxCharacterLimit, "The last name output should be the same length as the input")
	assert.IsType(t, "", *res, "The first name should be a string")
}

func Test_FirstNameTransformer(t *testing.T) {
	testVal := "bill"
	mapping := fmt.Sprintf(`root = transform_first_name(value:%q,preserve_length:true,max_length:%d)`, testVal, maxCharacterLimit)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the first name transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.NotNil(t, res, "The response shouldn't be nil.")

	resStr, ok := res.(*string)
	if !ok {
		t.Errorf("Expected *string, got %T", res)
		return
	}

	if resStr != nil {
		assert.Equal(t, len(*resStr), len(testVal), "Generated first name must be as long as input first name")
	} else {
		t.Error("Pointer is nil, expected a valid string pointer")
	}
}

func Test_TransformFirstNameTransformerWithEmptyValue(t *testing.T) {
	nilName := ""
	mapping := fmt.Sprintf(`root = transform_first_name(value:%q,preserve_length:true,max_length:%d)`, nilName, maxCharacterLimit)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the first name transformer")

	_, err = ex.Query(nil)
	assert.NoError(t, err)
}

func Test_TransformFirstNameTransformer_NoOptions(t *testing.T) {
	mapping := fmt.Sprintf(`root = transform_first_name(value:%q)`, "name")
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the first name transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
}
