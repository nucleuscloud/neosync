package neosync_transformers

import (
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func TestProcessGenderAbbreviateTrue(t *testing.T) {
	res, err := GenerateRandomGender(true)

	valid := []string{"f", "m", "u", "n"}

	assert.NoError(t, err)
	assert.Len(t, res, 1, "Generated gender must have a length of one")
	assert.Contains(t, valid, res, "Gender should be one of female, male, undefined, nonbinary")

}

func TestProcessGenderAbbreviateFalse(t *testing.T) {

	res, err := GenerateRandomGender(false)

	valid := []string{"female", "male", "undefined", "nonbinary"}

	assert.NoError(t, err)
	assert.Contains(t, valid, res, "Gender should be one of female, male, undefined, nonbinary")
}

func TestGenderTransformer(t *testing.T) {
	mapping := `root = gendertransformer(true)`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the gender transformer")

	res, err := ex.Query(nil)

	valid := []string{"f", "m", "u", "n"}

	assert.NoError(t, err)
	assert.Len(t, res, 1, "Generated gender must have a length of one")
	assert.Contains(t, valid, res, "Gender should be one of female, male, undefined, nonbinary")

}
