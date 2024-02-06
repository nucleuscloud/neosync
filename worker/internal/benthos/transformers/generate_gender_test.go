package transformers

import (
	"fmt"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

var maxGenderCharLimit = int64(6)

func Test_GenerateGenderAbbreviateTrue(t *testing.T) {
	res, err := GenerateRandomGender(true, maxGenderCharLimit)

	valid := []string{"f", "m", "u", "n"}

	assert.NoError(t, err)
	assert.Len(t, res, 1, "Generated gender must have a length of one")
	assert.Contains(t, valid, res, "Gender should be one of female, male, undefined, nonbinary")

}

func Test_GenerateGenderAbbreviateFalse(t *testing.T) {

	res, err := GenerateRandomGender(false, int64(20))

	valid := []string{"female", "male", "undefined", "nonbinary"}

	assert.NoError(t, err)
	assert.Contains(t, valid, res, "Gender should be one of female, male, undefined, nonbinary")
}

func Test_GenderTransformer(t *testing.T) {
	mapping := fmt.Sprintf(`root = generate_gender(abbreviate:true,max_length:%d)`, maxGenderCharLimit)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the gender transformer")

	res, err := ex.Query(nil)

	valid := []string{"f", "m", "u", "n"}

	assert.NoError(t, err)
	assert.Len(t, res, 1, "Generated gender must have a length of one")
	assert.Contains(t, valid, res, "Gender should be one of female, male, undefined, nonbinary")
	assert.Equal(t, int64(len(res.(string))), int64(1), "the length should be 1")

}
