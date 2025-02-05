package transformers

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"

	"github.com/redpanda-data/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func Test_TranformLastNameEmptyName(t *testing.T) {
	randomizer := rand.New(rand.NewSource(1))
	emptyName := ""

	res, err := transformLastName(randomizer, emptyName, true, maxCharacterLimit)
	assert.NoError(t, err)
	assert.Nil(t, res, "The response should be nil")
}

func Test_TransformLastName_Preserve_True(t *testing.T) {
	randomizer := rand.New(rand.NewSource(1))
	nameLength := int64(len(name))

	res, err := transformLastName(randomizer, name, true, maxCharacterLimit)

	assert.NoError(t, err)
	assert.Equal(t, nameLength, int64(len(*res)), "The last name output should be the same length as the input")
	assert.IsType(t, "", *res, "The last name should be a string")
}

func Test_TransformLastName_Preserve_True_With_Padding(t *testing.T) {
	randomizer := rand.New(rand.NewSource(1))
	length := 300

	var buffer bytes.Buffer
	char := "a"
	for i := 0; i < length; i++ {
		buffer.WriteString(char)
	}

	name := buffer.String()

	res, err := transformLastName(randomizer, name, true, 500)

	assert.NoError(t, err)
	assert.Equal(t, int64(len(name)), int64(len(*res)), "The last name output should be the same length as the input")
	assert.IsType(t, "", *res, "The last name should be a string")
}

func Test_TransformLastName_Preserve_False(t *testing.T) {
	randomizer := rand.New(rand.NewSource(1))

	res, err := transformLastName(randomizer, name, false, maxCharacterLimit)

	assert.NoError(t, err)
	assert.LessOrEqual(t, int64(len(*res)), maxCharacterLimit, "The last name output should be the same length as the input")
	assert.IsType(t, "", *res, "The last name should be a string")
}

func Test_LastNameTransformer(t *testing.T) {
	testVal := "bill"
	mapping := fmt.Sprintf(`root = transform_last_name(value:%q,preserve_length:true,max_length:%d)`, testVal, maxCharacterLimit)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the last name transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.NotNil(t, res, "The response shouldn't be nil.")

	resStr, ok := res.(*string)
	if !ok {
		t.Errorf("Expected *string, got %T", res)
		return
	}

	if resStr != nil {
		assert.Equal(t, len(*resStr), len(testVal), "Generated last name must be as long as input last name")
	} else {
		t.Error("Pointer is nil, expected a valid string pointer")
	}
}

func Test_TransformLastNameTransformerWithEmptyValue(t *testing.T) {
	nilName := ""
	mapping := fmt.Sprintf(`root = transform_last_name(value:%q,preserve_length:true,max_length:%d)`, nilName, maxCharacterLimit)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the last name transformer")

	_, err = ex.Query(nil)
	assert.NoError(t, err)
}

func Test_TransformLastNameTransformer_NoOptions(t *testing.T) {
	mapping := fmt.Sprintf(`root = transform_last_name(value:%q)`, "lastname")
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the last name transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
}
