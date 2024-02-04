package transformers

import (
	"fmt"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateE164FormatPhoneNumber(t *testing.T) {

	min := int64(9)
	max := int64(12)

	res, err := GenerateRandomE164PhoneNumber(min, max, maxCharacterLimit)

	assert.NoError(t, err)
	assert.Equal(t, ValidateE164(res), true, "The actual value should be a valid e164 number")
	assert.GreaterOrEqual(t, len(res), 9+1, "Should be greater than 10 characters in length. 9 for the number and 1 for the plus sign.")
	assert.LessOrEqual(t, len(res), 15+1, "Should be less than 16 characters in length. 15 for the number and 1 for the plus sign.")

}

func Test_GeneratePhoneNumberE164FormatPreserveLengthEqualMinMax(t *testing.T) {

	min := int64(12)
	max := int64(12)

	res, err := GenerateRandomE164PhoneNumber(min, max, maxCharacterLimit)

	assert.NoError(t, err)
	assert.Equal(t, ValidateE164(res), true, "The actual value should be a valid e164 number")
	assert.GreaterOrEqual(t, len(res), 8+1, "Should be greater than 9 characters in length. 9 for the number and 1 for the plus sign.")
	assert.LessOrEqual(t, len(res), 15+1, "Should be less than 16 characters in length. 15 for the number and 1 for the plus sign.")
	assert.Equal(t, int64(len(res)), max+1)

}

func Test_GeneratePhoneNumberE164FormatPreserveLengthShortMax(t *testing.T) {

	min := int64(9)
	max := int64(12)
	maxPhoneLimit := 11

	res, err := GenerateRandomE164PhoneNumber(min, max, int64(maxPhoneLimit))

	assert.NoError(t, err)
	assert.Equal(t, ValidateE164(res), true, "The actual value should be a valid e164 number")
	assert.GreaterOrEqual(t, len(res), 8+1, "Should be greater than 9 characters in length. 9 for the number and 1 for the plus sign.")
	assert.LessOrEqual(t, len(res), maxPhoneLimit+1, "Should be less than 16 characters in length. 15 for the number and 1 for the plus sign.")

}

func Test_GenerateE164PhoneNumberTransformer(t *testing.T) {

	min := int64(10)
	max := int64(13)
	mapping := fmt.Sprintf(`root = generate_e164_phone_number(min:%d,max:%d,max_length:%d)`, min, max, maxCharacterLimit)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the phone transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)
	assert.Equal(t, ValidateE164(res.(string)), true, "The actual value should be a valid e164 number")
	assert.GreaterOrEqual(t, len(res.(string)), 8+1, "Should be greater than 9 characters in length. 9 for the number and 1 for the plus sign.")
	assert.LessOrEqual(t, len(res.(string)), 15+1, "Should be less than 16 characters in length. 15 for the number and 1 for the plus sign.")
}

func Test_ValidateE164True(t *testing.T) {

	val := "+6272636472"

	res := ValidateE164(val)

	assert.Equal(t, res, true, "The e164 number should have a plus sign at the 0th index and be 9 < x < 15 characters long.")
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

	assert.Equal(t, res, false, "The e164 number should  be 9 < x")
}
