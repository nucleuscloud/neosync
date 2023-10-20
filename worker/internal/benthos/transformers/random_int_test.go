package neosync_transformers

import (
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func TestProcessRandomIntPreserveLengthTrue(t *testing.T) {

	val := int64(67543543)
	expectedLength := int64(8)

	res, err := ProcessRandomInt(val, true, 0)

	assert.NoError(t, err)
	assert.Equal(t, GetIntLength(res), expectedLength, "The output int needs to be the same length as the input int")

}

func TestProcessRandomIntPreserveLengthFalse(t *testing.T) {

	val := int64(67543543)
	expectedLength := int64(4)

	res, err := ProcessRandomInt(val, false, expectedLength)

	assert.NoError(t, err)
	assert.Equal(t, GetIntLength(res), expectedLength, "The output int needs to be the same length as the input int")

}

func TestProcessRandomIntPreserveLengthTrueIntLength(t *testing.T) {

	val := int64(67543543)
	expectedLength := int64(8)

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

// //not valid
// the count:  8
// min 10000000 //this is the smallest value that still has 8 integers
// min length 8
// max 100000000 //this is the upper bound but is exclusive so this should be 9 integers
// max length 9
// rand 97367785 //this is the random int before the min gets added, this should be between the min and max
// res 107367785 //this adds the min to the rand so that it's 9 which in this case is wrong,
// res length 9
// res does not equal count

// //valid
// the count:  8
// min 10000000 //this count is still good
// min length 8
// max 100000000 //this max is good
// max length 9
// rand 5507875
// res 15507875
// res length 8
// res equals count

// //the issue is taht the rand sometimes generates a value that has 7 when it shoudl generate a value that
// // should always have 8, the rand.Int generates a val from 0 -> Max, so it could generate a val that only ius onl
// // 3 value long

// //the edge case is when the rand generates a value that is count long and starts with a 9 it pushes it over when it adds the min which pushes it to
// // the count + 1 and is an error, so it had a 1/10 chance of failing
