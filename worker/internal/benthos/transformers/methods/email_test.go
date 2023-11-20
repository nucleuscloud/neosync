package neosync_benthos_transformers_methods

import (
	"regexp"
	"strings"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

const email = "evis@gmail.com"

func TestGenerateEmail(t *testing.T) {

	res, err := GenerateEmailPreserveDomain(email, true)

	assert.NoError(t, err)
	/* There is a very small chance that the randomly generated email address actually matches
	the input email address which is why can't do an assert.NoEqual() but instead just have to check
	that the email has the correct structrue */
	assert.Equal(t, true, isValidEmail(res), "true", "The domain should not explicitly be preserved but randomly generated.")
}

func TestGenerateEmailPreserveDomain(t *testing.T) {

	res, err := GenerateEmailPreserveDomain(email, true)

	assert.NoError(t, err)
	/* There is a very small chance that the randomly generated email address actually matches
	the input email address which is why can't do an assert.NoEqual() but instead just have to check
	that the email has the correct structrue */
	assert.Equal(t, true, isValidEmail(res), "true", "The domain should not explicitly be preserved but randomly generated.")
}

func TestGenerateEmailPreserveLength(t *testing.T) {

	res, err := GenerateEmailPreserveLength(email, true)

	assert.NoError(t, err)
	assert.Equal(t, len(email), len(res), "The length of the emails should be the same")
}

func TestGenerateEmailPreserveLengthTruePreserveDomainTrue(t *testing.T) {
	email := "johndoe@gmail.com"

	res, err := GenerateEmailPreserveDomainAndLength(email, true, true)

	assert.NoError(t, err)
	assert.Equal(t, true, isValidEmail(res), "The expected email should be have a valid email structure")

}

func TestGenerateEmailPreserveLengthFalsePreserveDomainFalse(t *testing.T) {
	email := "johndoe@gmail.com"

	res, err := GenerateEmail(email, false, false)

	assert.NoError(t, err)
	assert.Equal(t, true, isValidEmail(res), "The expected email should be have a valid email structure")

}

func TestGenerateDomain(t *testing.T) {

	res, err := GenerateDomain()
	assert.NoError(t, err)

	assert.Equal(t, true, IsValidDomain(res))

}

func TestGenerateUsername(t *testing.T) {

	res, err := GenerateRandomUsername()
	assert.NoError(t, err)

	assert.Equal(t, true, IsValidUsername(res))

}

func TestEmailTransformer(t *testing.T) {
	mapping := `root = this.emailtransformer(true, true)`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the email transformer")

	testVal := "evil@gmail.com"

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

func IsValidDomain(domain string) bool {
	pattern := `^@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

	// Compile the regex pattern
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}

	// Use the regex pattern to validate the email
	return re.MatchString(domain)
}

func IsValidUsername(domain string) bool {
	pattern := `^[a-zA-Z0-9]`

	// Compile the regex pattern
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}

	// Use the regex pattern to validate the email
	return re.MatchString(domain)
}
