package transformers

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/stretchr/testify/assert"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

func Test_GenerateInternationalPhoneNumber(t *testing.T) {
	minValue := int64(9)
	maxValue := int64(12)

	res, err := generateInternationalPhoneNumber(rng.New(time.Now().UnixNano()), minValue, maxValue)

	assert.NoError(t, err)
	assert.Equal(t, validateE164(res), true, "The actual value should be a valid e164 number")
	assert.GreaterOrEqual(t, len(res), 9, "Should be greater than 10 characters in length. 9 for the number and 1 for the plus sign.")
	assert.LessOrEqual(t, len(res), 15, "Should be less than 16 characters in length. 15 for the number and 1 for the plus sign.")
}

func Test_GenerateInternationalPhoneNumberPreserveLength(t *testing.T) {
	minValue := int64(12)
	maxValue := int64(12)

	res, err := generateInternationalPhoneNumber(rng.New(time.Now().UnixNano()), minValue, maxValue)

	assert.NoError(t, err)
	assert.Equal(t, validateE164(res), true, "The actual value should be a valid e164 number")
	assert.GreaterOrEqual(t, len(res), 9+1, "Should be greater than 10 characters in length. 9 for the number and 1 for the plus sign.")
	assert.LessOrEqual(t, len(res), 15+1, "Should be less than 16 characters in length. 15 for the number and 1 for the plus sign.")
}

func Test_GenerateInternationalPhoneNumberTransformer(t *testing.T) {
	minValue := int64(10)
	maxValue := int64(13)
	mapping := fmt.Sprintf(`root = generate_e164_phone_number(min:%d, max: %d)`, minValue, maxValue)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the international phone number transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)
	assert.Equal(t, validateE164(res.(string)), true, "The actual value should be a valid e164 number")
	assert.GreaterOrEqual(t, len(res.(string)), 9+1, "Should be greater than 10 characters in length. 9 for the number and 1 for the plus sign.")
	assert.LessOrEqual(t, len(res.(string)), 15+1, "Should be less than 16 characters in length. 15 for the number and 1 for the plus sign.")
}

func Test_GenerateInternationalPhoneNumberTransformer_NoOptions(t *testing.T) {
	mapping := `root = generate_e164_phone_number()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the international phone number transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
}

func Test_ValidateE164True(t *testing.T) {
	val := "+6272636472"

	res := validateE164(val)

	assert.Equal(t, res, true, "The e164 number should have a plus sign at the 0th index and be 10 < x < 15 characters long.")
}

func Test_ValidateE164FalseTooLong(t *testing.T) {
	val := "627263647278439"

	res := validateE164(val)

	assert.Equal(t, res, false, "The e164 number should  be x < 15 characters long.")
}

func Test_ValidateE164FalseNoPlusSign(t *testing.T) {
	val := "6272636472784"

	res := validateE164(val)

	assert.Equal(t, res, false, "The e164 number should have a plus sign at the beginning.")
}

func Test_ValidateE164FalseTooshort(t *testing.T) {
	val := "627263"

	res := validateE164(val)

	assert.Equal(t, res, false, "The e164 number should  be 10 < x")
}

func validateE164(p string) bool {
	if len(p) >= 10 && len(p) <= 15 && strings.Contains(p, "+") {
		return true
	}
	return false
}
