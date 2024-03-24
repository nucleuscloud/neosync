package transformers

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func Test_TranformLastNameEmptyName(t *testing.T) {
	randomizer := rand.New(rand.NewSource(1))
	emptyName := ""

	res, err := transformLastName(randomizer, emptyName, true, maxCharacterLimit)
	assert.NoError(t, err)
	assert.Nil(t, res, "The response should be nil")
}

func Test_TransformLastNamePreserveLengthTrue(t *testing.T) {
	randomizer := rand.New(rand.NewSource(1))
	nameLength := int64(len(name))

	res, err := transformLastName(randomizer, name, true, maxCharacterLimit)

	assert.NoError(t, err)
	assert.Equal(t, nameLength, int64(len(*res)), "The last name output should be the same length as the input")
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
