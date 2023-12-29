package transformers

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateSSN(t *testing.T) {

	for i := 0; i < 100; i++ {
		res, err := GenerateRandomSSN()
		assert.NoError(t, err)
		fmt.Println("res", res)
	}

	// assert.NoError(t, err)
	// assert.IsType(t, "", res, "The actual value type should be a string")
	// assert.True(t, isValidSSN(res), `The returned ssn should follow this regex = ^\d{3}-\d{2}-\d{4}$)`)
}

func Test_SSNTransformer(t *testing.T) {
	mapping := `root = generate_ssn()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the ssn transformer")

	res, err := ex.Query(nil)

	assert.NoError(t, err)
	assert.IsType(t, "", res, "The actual value type should be a string")
	assert.True(t, isValidSSN(res.(string)), `The returned ssn should follow this regex = ^\d{3}-\d{2}-\d{4}$)`)
}

func isValidSSN(ssn string) bool {
	regex := regexp.MustCompile(`^\d{3}-\d{2}-\d{4}$`)
	return regex.MatchString(ssn)
}
