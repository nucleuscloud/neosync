package transformers

import (
	"fmt"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateInternationalPhoneNumber(t *testing.T) {
	min := int64(9)
	max := int64(12)

	res, err := GenerateInternationalPhoneNumber(min, max)

	assert.NoError(t, err)
	assert.Equal(t, ValidateE164(res), true, "The actual value should be a valid e164 number")
	assert.GreaterOrEqual(t, len(res), 9, "Should be greater than 10 characters in length. 9 for the number and 1 for the plus sign.")
	assert.LessOrEqual(t, len(res), 15, "Should be less than 16 characters in length. 15 for the number and 1 for the plus sign.")
}

func Test_GenerateInternationalPhoneNumberPreserveLength(t *testing.T) {
	min := int64(12)
	max := int64(12)

	res, err := GenerateInternationalPhoneNumber(min, max)

	assert.NoError(t, err)
	assert.Equal(t, ValidateE164(res), true, "The actual value should be a valid e164 number")
	assert.GreaterOrEqual(t, len(res), 9+1, "Should be greater than 10 characters in length. 9 for the number and 1 for the plus sign.")
	assert.LessOrEqual(t, len(res), 15+1, "Should be less than 16 characters in length. 15 for the number and 1 for the plus sign.")
}

func Test_GenerateInternationalPhoneNumberTransformer(t *testing.T) {
	min := int64(10)
	max := int64(13)
	mapping := fmt.Sprintf(`root = generate_e164_phone_number(min:%d, max: %d)`, min, max)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the international phone number transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)
	assert.Equal(t, ValidateE164(res.(string)), true, "The actual value should be a valid e164 number")
	assert.GreaterOrEqual(t, len(res.(string)), 9+1, "Should be greater than 10 characters in length. 9 for the number and 1 for the plus sign.")
	assert.LessOrEqual(t, len(res.(string)), 15+1, "Should be less than 16 characters in length. 15 for the number and 1 for the plus sign.")
}

func Test_ValidateE164True(t *testing.T) {
	val := "+6272636472"

	res := ValidateE164(val)

	assert.Equal(t, res, true, "The e164 number should have a plus sign at the 0th index and be 10 < x < 15 characters long.")
}

func Test_ValidateE164FalseTooLong(t *testing.T) {
	val := "627263647278439"

	res := ValidateE164(val)

	assert.Equal(t, res, false, "The e164 number should  be x < 15 characters long.")
}

func Test_ValidateE164FalseNoPlusSign(t *testing.T) {
	val := "6272636472784"

	res := ValidateE164(val)

	assert.Equal(t, res, false, "The e164 number should have a plus sign at the beginning.")
}

func Test_ValidateE164FalseTooshort(t *testing.T) {
	val := "627263"

	res := ValidateE164(val)

	assert.Equal(t, res, false, "The e164 number should  be 10 < x")
}
