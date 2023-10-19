package neosync_transformers

import (
	"regexp"
	"strings"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func TestProcessEmailPreserveLengthTrue(t *testing.T) {

	email := "evis@gmail.com"

	res, err := ProcessEmail(email, true, true)

	assert.NoError(t, err)
	assert.Equal(t, len(res), len(email), "The length of the emails should be the same")
}

func TestProcessEmailPreserveLengthFalse(t *testing.T) {
	email := "johndoe@gmail.com"

	res, err := ProcessEmail(email, false, true)

	assert.NoError(t, err)
	assert.Equal(t, true, isValidEmail(res), "The expected email should be have a valid email structure")
}

func TestProcessEmailPreserveDomain(t *testing.T) {

	email := "evis@gmail.com"

	res, err := ProcessEmail(email, true, true)

	assert.NoError(t, err)
	assert.Equal(t, strings.Split(res, "@")[1], "gmail.com", "The domain of the input email should be the same as the output email")

}

func TestProcessEmailNoPreserveDomain(t *testing.T) {

	email := "evis@gmail.com"

	res, err := ProcessEmail(email, true, false)

	assert.NoError(t, err)
	/* There is a very small chance that the randomly generated email address actually matches
	the input email address which is why can't do an assert.NoEqual() but instead just have to check
	that the email has the correct structrue */
	assert.Equal(t, true, isValidEmail(res), "true", "The domain should not explicitly be preserved but randomly generated.")

}

func TestProcessEmailPreserveLengthFalsePreserveDomainFalse(t *testing.T) {
	email := "johndoe@gmail.com"

	res, err := ProcessEmail(email, false, false)

	assert.NoError(t, err)
	assert.Equal(t, true, isValidEmail(res), "The expected email should be have a valid email structure")

}

func TestEmailTransformer(t *testing.T) {
	mapping := `root = this.emailtransformer(true, true)`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the email transformer")

	testVal := "evis@gmail.com"

	res, err := ex.Query(testVal)
	assert.NoError(t, err)

	assert.Len(t, res.(string), len(testVal), "Generated email must be the same length as the input email")
	assert.Equal(t, strings.Split(res.(string), "@")[1], "gmail.com")
}

func isValidEmail(email string) bool {
	// Regular expression pattern for a simple email validation
	emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	regex := regexp.MustCompile(emailPattern)
	return regex.MatchString(email)
}
