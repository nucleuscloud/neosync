package neosync_transformers

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const email = "evis@gmail.com"

func Test_GenerateEmail(t *testing.T) {

	res, err := GenerateEmailPreserveDomain(email, true)

	assert.NoError(t, err)
	/* There is a very small chance that the randomly generated email address actually matches
	the input email address which is why can't do an assert.NoEqual() but instead just have to check
	that the email has the correct structrue */
	assert.Equal(t, true, isValidEmail(res), "true", "The domain should not explicitly be preserved but randomly generated.")
}

func Test_GenerateEmailPreserveDomain(t *testing.T) {

	res, err := GenerateEmailPreserveDomain(email, true)

	assert.NoError(t, err)
	/* There is a very small chance that the randomly generated email address actually matches
	the input email address which is why can't do an assert.NoEqual() but instead just have to check
	that the email has the correct structrue */
	assert.Equal(t, true, isValidEmail(res), "true", "The domain should not explicitly be preserved but randomly generated.")
}

func Test_GenerateEmailPreserveLength(t *testing.T) {

	res, err := GenerateEmailPreserveLength(email, true)

	assert.NoError(t, err)
	assert.Equal(t, len(email), len(res), "The length of the emails should be the same")
}

func Test_GenerateEmailPreserveLengthTruePreserveDomainTrue(t *testing.T) {
	email := "johndoe@gmail.com"

	res, err := GenerateEmailPreserveDomainAndLength(email, true, true)

	assert.NoError(t, err)
	assert.Equal(t, true, isValidEmail(res), "The expected email should be have a valid email structure")

}

func Test_GenerateEmailPreserveLengthFalsePreserveDomainFalse(t *testing.T) {
	email := "johndoe@gmail.com"

	res, err := GenerateEmail(email, false, false)

	assert.NoError(t, err)
	assert.Equal(t, true, isValidEmail(res), "The expected email should be have a valid email structure")

}

func Test_GenerateDomain(t *testing.T) {

	res, err := GenerateDomain()
	assert.NoError(t, err)

	assert.Equal(t, true, IsValidDomain(res), "The expected email should have a valid domain")

}

func Test_GenerateUsername(t *testing.T) {

	res, err := GenerateRandomUsername()
	assert.NoError(t, err)

	assert.Equal(t, true, IsValidUsername(res), "The expected email should have a valid username")

}

func Test_GenerateRandomEmail(t *testing.T) {
	res, err := GenerateRandomEmail()

	assert.NoError(t, err)
	assert.Equal(t, true, isValidEmail(res), "The expected email should be have a valid email structure")
}

func Test_EmailTransformerWithValue(t *testing.T) {
	testVal := "evil@gmail.com"
	mapping := fmt.Sprintf(`root = emailtransformer(%q,true,true)`, testVal)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err)
	assert.NoError(t, err, "failed to parse the email transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.Len(t, res.(string), len(testVal), "Generated email must be the same length as the input email")
	assert.Equal(t, strings.Split(res.(string), "@")[1], "gmail.com")
}

func Test_EmailTransformerWithEmptyValue(t *testing.T) {
	testVal := ""
	mapping := fmt.Sprintf(`root = emailtransformer(%q,true,true)`, testVal)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err)
	assert.NoError(t, err, "failed to parse the email transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.Equal(t, true, isValidEmail(res.(string)))
}

func Test_EmailTransformerEmailParamError(t *testing.T) {
	mapping := `root = emailtransformer(,true,true)`
	_, err := bloblang.Parse(mapping)
	assert.Error(t, err, "failed to parse the email transformer, missing param")

}
func Test_EmailTransformerPreserveLengthParamError(t *testing.T) {
	testVal := ""
	mapping := fmt.Sprintf(`root = emailtransformer(%q,true)`, testVal)
	_, err := bloblang.Parse(mapping)
	assert.Error(t, err, "failed to parse the email transformer, missing param")

}
func Test_EmailTransformerErrorParams(t *testing.T) {
	mapping := `root = emailtransformer(,true,true)`
	_, err := bloblang.Parse(mapping)
	assert.Error(t, err, "failed to parse the email transformer, missing param")

}

func Test_ParseEmailError(t *testing.T) {
	test := "ehiu.com"

	_, err := parseEmail(test)
	assert.Error(t, err)

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

type MockParsedParams struct {
	mock.Mock
}

func (m *MockParsedParams) GetString(key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

func (m *MockParsedParams) GetBool(key string) (bool, error) {
	args := m.Called(key)
	return args.Bool(0), args.Error(1)
}
