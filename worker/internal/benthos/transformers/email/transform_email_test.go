package transformers_email

import (
	"fmt"
	"strings"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

const email = "evis@gmail.com"

func Test_GenerateEmailPreserveLengthFalsePreserveDomainTrue(t *testing.T) {

	res, err := TransformEmail(email, false, true)

	assert.NoError(t, err)
	assert.Equal(t, true, IsValidEmail(res), "The expected email should be have a valid email structure")
	assert.Equal(t, "gmail.com", strings.Split(res, "@")[1])

}

func Test_GenerateEmailPreserveLengthFalsePreserveDomainFalse(t *testing.T) {

	res, err := TransformEmail(email, false, false)

	assert.NoError(t, err)
	assert.Equal(t, true, IsValidEmail(res), "The expected email should be have a valid email structure")

}

func Test_GenerateEmailPreserveLengthTruePreserveDomainFalse(t *testing.T) {

	res, err := TransformEmail(email, true, false)

	assert.NoError(t, err)
	assert.Equal(t, len(email), len(res), "The expected email should be have a valid email structure")

}

func Test_GenerateEmail(t *testing.T) {

	res, err := TransformEmailPreserveDomain(email, true)

	assert.NoError(t, err)
	/* There is a very small chance that the randomly generated email address actually matches
	the input email address which is why can't do an assert.NoEqual() but instead just have to check
	that the email has the correct structrue */
	assert.Equal(t, true, IsValidEmail(res), "true", "The domain should not explicitly be preserved but randomly generated.")
}

func Test_GenerateEmailPreserveDomain(t *testing.T) {

	res, err := TransformEmailPreserveDomain(email, true)

	assert.NoError(t, err)
	assert.Equal(t, true, IsValidEmail(res), "true", "The domain should not explicitly be preserved but randomly generated.")
}

func Test_GenerateEmailPreserveLength(t *testing.T) {

	res, err := TransformEmailPreserveLength(email, true)

	assert.NoError(t, err)
	assert.Equal(t, len(email), len(res), "The length of the emails should be the same")
}

func Test_GenerateEmailPreserveLengthTruePreserveDomainTrue(t *testing.T) {

	res, err := TransformEmailPreserveDomainAndLength(email, true, true)

	assert.NoError(t, err)
	assert.Equal(t, true, IsValidEmail(res), "The expected email should be have a valid email structure")

}

func Test_GenerateUsername(t *testing.T) {

	res, err := GenerateRandomUsername()
	assert.NoError(t, err)

	assert.Equal(t, true, IsValidUsername(res), "The expected email should have a valid username")

}

func Test_EmailTransformerWithValue(t *testing.T) {
	mapping := fmt.Sprintf(`root = transform_email(%q,true,true)`, email)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the email transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.Len(t, res.(string), len(email), "Generated email must be the same length as the input email")
	assert.Equal(t, strings.Split(res.(string), "@")[1], "The domain name should be gmail.com")
}
