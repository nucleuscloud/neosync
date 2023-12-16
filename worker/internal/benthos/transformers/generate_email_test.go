package transformers

import (
	"fmt"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateRandomEmail(t *testing.T) {

	min := int64(3)
	max := int64(7)

	res, err := GenerateRandomEmail(min, max)

	assert.NoError(t, err)
	assert.Equal(t, true, transformer_utils.IsValidEmail(res), "The expected email should be have a valid email format")
	assert.GreaterOrEqual(t, len(res), min, "the string should be greater than or equal to the min value")
	assert.LessOrEqual(t, len(res), max, "the string should be less than or equal to the max value")

}

func Test_GenerateRandomDomain(t *testing.T) {

	min := int64(3)
	max := int64(7)

	res, err := GenerateRandomDomain(min, max)
	assert.NoError(t, err)

	assert.Equal(t, true, transformer_utils.IsValidDomain(res), "The expected email should have a valid domain")
	assert.GreaterOrEqual(t, len(res), min, "the string should be greater than or equal to the min value")
	assert.LessOrEqual(t, len(res), max, "the string should be less than or equal to the max value")

}

func Test_GenerateRandomUsername(t *testing.T) {

	min := int64(3)
	max := int64(7)

	res, err := GenerateRandomUsername(min, max)
	assert.NoError(t, err)

	assert.Equal(t, true, transformer_utils.IsValidUsername(res), "The expected email should have a valid username")
	assert.GreaterOrEqual(t, len(res), min, "the string should be greater than or equal to the min value")
	assert.LessOrEqual(t, len(res), max, "the string should be less than or equal to the max value")

}

func Test_RandomEmailTransformer(t *testing.T) {

	min := int64(3)
	max := int64(7)

	mapping := fmt.Sprintf(`root = generate_email(min:%d, %d)`, min, max)
	ex, err := bloblang.Parse(mapping)

	assert.NoError(t, err, "failed to parse the email transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.Equal(t, true, transformer_utils.IsValidEmail(res.(string)), " The expected email should have a valid format")
	assert.GreaterOrEqual(t, len(res.(string)), min, "the string should be greater than or equal to the min value")
	assert.LessOrEqual(t, len(res.(string)), max, "the string should be less than or equal to the max value")

}
