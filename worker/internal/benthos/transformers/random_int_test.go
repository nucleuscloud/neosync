package neosync_transformers

import (
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func TestProcessRandomIntPreserveLengthTrue(t *testing.T) {

	val := int64(67543543)
	expectedLength := 8

	res, err := ProcessRandomInt(val, true, 0)

	assert.NoError(t, err)
	assert.Equal(t, GetIntLength(res), expectedLength, "The output int needs to be the same length as the input int")

}

func TestProcessRandomIntPreserveLengthFalse(t *testing.T) {

	val := int64(67543543)
	expectedLength := 4

	res, err := ProcessRandomInt(val, false, int64(expectedLength))

	assert.NoError(t, err)
	assert.Equal(t, GetIntLength(res), expectedLength, "The output int needs to be the same length as the input int")

}

func TestProcessRandomIntPreserveLengthTrueIntLength(t *testing.T) {

	val := int64(67543543)
	expectedLength := 8

	res, err := ProcessRandomInt(val, true, int64(5))

	assert.NoError(t, err)
	assert.Equal(t, GetIntLength(res), expectedLength, "The output int needs to be the same length as the input int")

}

func TestRandomIntTransformer(t *testing.T) {
	mapping := `root = this.randominttransformer(true, 6)`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the random int transformer")

	testVal := int64(397283)

	res, err := ex.Query(testVal)
	assert.NoError(t, err)

	assert.Equal(t, GetIntLength(testVal), GetIntLength(res.(int64))) // Generated int must be the same length as the input int"
	assert.IsType(t, res, testVal)
}
