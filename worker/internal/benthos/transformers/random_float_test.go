package neosync_transformers

import (
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func TestProcessRandomFloatPreserveLength(t *testing.T) {

	val := float64(6754.3543)
	expectedLength := 8

	res, err := ProcessRandomFloat(val, true, int64(3), int64(3))

	actual := GetFloatLength(res).DigitsBeforeDecimalLength + GetFloatLength(res).DigitsAfterDecimalLength

	assert.NoError(t, err)
	assert.Equal(t, expectedLength, actual, "The output float needs to be the same length as the input Float")

}

func TestProcessRandomFloatPreserveLengthFalse(t *testing.T) {

	val := float64(6754.3543)
	expectedLength := 5

	res, err := ProcessRandomFloat(val, false, int64(3), int64(3))

	actual := GetFloatLength(res).DigitsAfterDecimalLength + GetFloatLength(res).DigitsBeforeDecimalLength
	assert.NoError(t, err)
	assert.Equal(t, actual, expectedLength, "The output Float needs to be the same length as the input Float")
}

func TestRandomFloatTransformer(t *testing.T) {
	mapping := `root = this.randomfloattransformer(true, 2,3)`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the random float transformer")

	testVal := float64(397.34)

	res, err := ex.Query(testVal)
	assert.NoError(t, err)

	assert.Equal(t, GetFloatLength(testVal), GetFloatLength(res.(float64))) // Generated Float must be the same length as the input Float"
	assert.IsType(t, res, testVal)
}
