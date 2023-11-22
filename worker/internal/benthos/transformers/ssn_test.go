package transformers

import (
	"regexp"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func TestGenerateSSN(t *testing.T) {

	res, err := GenerateRandomSSN()

	assert.NoError(t, err)
	assert.IsType(t, "", res)
	assert.True(t, isValidSSN(res), `The returned ssn should follow this regex = ^\d{3}-\d{2}-\d{4}$)`)
}

func TestSSNTransformer(t *testing.T) {
	mapping := `root = ssntransformer()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the ssn transformer")

	res, err := ex.Query(nil)

	assert.NoError(t, err)
	assert.IsType(t, "", res)
	assert.True(t, isValidSSN(res.(string)), `The returned ssn should follow this regex = ^\d{3}-\d{2}-\d{4}$)`)
}

func isValidSSN(ssn string) bool {
	regex := regexp.MustCompile(`^\d{3}-\d{2}-\d{4}$`)
	return regex.MatchString(ssn)
}
