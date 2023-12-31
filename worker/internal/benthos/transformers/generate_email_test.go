package transformers

import (
	"fmt"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateRandomEmail(t *testing.T) {

	res, err := GenerateEmail()

	assert.NoError(t, err)
	assert.Equal(t, true, transformer_utils.IsValidEmail(res), fmt.Sprintf(`The expected email should be have a valid email format. Received:%s`, res))
}

func Test_GenerateRandomDomain(t *testing.T) {

	res, err := GenerateEmailDomain()
	assert.NoError(t, err)

	assert.Equal(t, true, transformer_utils.IsValidDomain(res), "The expected email should have a valid domain")

}

func Test_GenerateRandomUsername(t *testing.T) {

	res, err := GenerateEmailUsername()
	assert.NoError(t, err)

	assert.Equal(t, true, transformer_utils.IsValidUsername(res), "The expected email should have a valid username")

}

func Test_RandomEmailTransformer(t *testing.T) {
	mapping := `root = generate_email()`
	ex, err := bloblang.Parse(mapping)

	assert.NoError(t, err, "failed to parse the email transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.Equal(t, true, transformer_utils.IsValidEmail(res.(string)), "The expected email should have a valid email format")
}
