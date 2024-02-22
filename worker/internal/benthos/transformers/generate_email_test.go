package transformers

import (
	"fmt"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateRandomEmailShort(t *testing.T) {
	shortMaxLength := int64(14)

	res, err := GenerateRandomEmail(shortMaxLength)

	assert.NoError(t, err)
	assert.Equal(t, true, transformer_utils.IsValidEmail(res), fmt.Sprintf(`The expected email should be have a valid email format. Received:%s`, res))
	assert.LessOrEqual(t, int64(len(res)), shortMaxLength, fmt.Sprintf("The city should be less than or equal to the max length. This is the error city:%s", res))
}

func Test_GenerateRandomEmail(t *testing.T) {
	for i := 0; i < 100; i++ {

		res, err := GenerateRandomEmail(int64(40))

		fmt.Println("res", res)

		assert.NoError(t, err)
		assert.Equal(t, true, transformer_utils.IsValidEmail(res), fmt.Sprintf(`The expected email should be have a valid email format. Received:%s`, res))
		assert.LessOrEqual(t, int64(len(res)), int64(40), fmt.Sprintf("The city should be less than or equal to the max length. This is the error city:%s", res))
	}
}

func Test_RandomEmailTransformer(t *testing.T) {
	maxLength := int64(40)
	mapping := fmt.Sprintf(`root = generate_email(max_length:%d)`, maxLength)
	ex, err := bloblang.Parse(mapping)

	assert.NoError(t, err, "failed to parse the email transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.LessOrEqual(t, int64(len(res.(string))), maxLength, fmt.Sprintf("The email should be less than or equal to the max length. This is the error email:%s", res))

	assert.Equal(t, true, transformer_utils.IsValidEmail(res.(string)), "The expected email should have a valid email format")
}
