package transformers_email

import (
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateRealisticEmail(t *testing.T) {

	res, err := GenerateRealisticEmail()

	assert.NoError(t, err)
	assert.Equal(t, true, IsValidEmail(res), "The expected email should be have a valid email format")

}

func Test_GenerateRealisticDomain(t *testing.T) {

	res, err := GenerateRealisticDomain()
	assert.NoError(t, err)

	assert.Equal(t, true, IsValidDomain(res), "The expected email should have a valid domain")

}

func Test_GenerateRealisticUsername(t *testing.T) {

	res, err := GenerateRealisticUsername()
	assert.NoError(t, err)

	assert.Equal(t, true, IsValidUsername(res), "The expected email should have a valid username")

}

func Test_RealisticmailTransformer(t *testing.T) {
	mapping := `root = generate_realistic_email()`
	ex, err := bloblang.Parse(mapping)

	assert.NoError(t, err, "failed to parse the email transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.Equal(t, true, IsValidEmail(res.(string)), "The expected email should have a valid email format")
}
