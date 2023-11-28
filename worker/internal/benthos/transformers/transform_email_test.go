package transformers

import (
	"fmt"
	"strings"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
	"github.com/stretchr/testify/assert"
)

const email = "evis@gmail.com"

func Test_GenerateEmailPreserveLengthFalsePreserveDomainTrue(t *testing.T) {

	res, err := TransformEmail(email, false, true)

	assert.NoError(t, err)
	assert.Equal(t, true, transformer_utils.IsValidEmail(*res), "The expected email should be have a valid email structure")
	assert.Equal(t, "gmail.com", strings.Split(*res, "@")[1])

}

func Test_GenerateEmailPreserveLengthFalsePreserveDomainFalse(t *testing.T) {

	res, err := TransformEmail(email, false, false)

	assert.NoError(t, err)
	assert.Equal(t, true, transformer_utils.IsValidEmail(*res), "The expected email should be have a valid email structure")

}

func Test_GenerateEmailPreserveLengthTruePreserveDomainFalse(t *testing.T) {

	res, err := TransformEmail(email, true, false)

	assert.NoError(t, err)
	assert.Equal(t, len(email), len(*res), "The expected email should be have a valid email structure")

}

func Test_GenerateEmail(t *testing.T) {

	res, err := TransformEmailPreserveDomain(email, true)

	assert.NoError(t, err)
	/* There is a very small chance that the randomly generated email address actually matches
	the input email address which is why can't do an assert.NoEqual() but instead just have to check
	that the email has the correct structrue */
	assert.Equal(t, true, transformer_utils.IsValidEmail(res), "true", "The domain should not explicitly be preserved but randomly generated.")
}

func Test_GenerateEmailPreserveDomain(t *testing.T) {

	res, err := TransformEmailPreserveDomain(email, true)

	assert.NoError(t, err)
	assert.Equal(t, true, transformer_utils.IsValidEmail(res), "true", "The domain should not explicitly be preserved but randomly generated.")
}

func Test_GenerateEmailPreserveLength(t *testing.T) {

	res, err := TransformEmailPreserveLength(email, true)

	assert.NoError(t, err)
	assert.Equal(t, len(email), len(res), "The length of the emails should be the same")
}

func Test_GenerateEmailPreserveLengthTruePreserveDomainTrue(t *testing.T) {

	res, err := TransformEmailPreserveDomainAndLength(email, true, true)

	assert.NoError(t, err)
	assert.Equal(t, true, transformer_utils.IsValidEmail(res), "The expected email should be have a valid email structure")

}

func Test_GenerateEmailUsername(t *testing.T) {

	res, err := GenerateRandomUsername()
	assert.NoError(t, err)

	assert.Equal(t, true, transformer_utils.IsValidUsername(res), "The expected email should have a valid username")

}

func Test_TransformEmailTransformerWithValue(t *testing.T) {
	mapping := fmt.Sprintf(`root = transform_email(value:%q,preserve_domain:true,preserve_length:true)`, email)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the email transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.NotNil(t, res, "The response shouldn't be nil.")

	resStr, ok := res.(*string)
	if !ok {
		t.Errorf("Expected *string, got %T", res)
		return
	}

	if resStr != nil {
		assert.Equal(t, len(*resStr), len(email), "Generated email must be the same length as the input email")
		assert.Equal(t, strings.Split(*resStr, "@")[1], "gmail.com", "The actual value should be have gmail.com as the domain")
	} else {
		t.Error("Pointer is nil, expected a valid string pointer")
	}
}

// the case where the input value is null
func Test_TransformEmailTransformerWithEmptyValue(t *testing.T) {

	nilEmail := ""
	mapping := fmt.Sprintf(`root = transform_email(value:%q,preserve_domain:true,preserve_length:true)`, nilEmail)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the email transformer")

	_, err = ex.Query(nil)
	assert.NoError(t, err)
}
