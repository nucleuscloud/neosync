package neosync_transformers

import (
	"strings"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func TestProcessPhoneNumberPreserveLengthTrue(t *testing.T) {

	pn := "1838492832"
	expectedLength := 10

	res, err := ProcessPhoneNumber(pn, true, false, false)

	assert.NoError(t, err)
	assert.Equal(t, len(res), expectedLength, "The length of the output phone number should be the same as the input phone number")

}

func TestProcessPhoneNumberPreserveLengthTrueHyphens(t *testing.T) {

	// we strip the hyphens when we process the phone number and the include hyphens param is set to false so the return val will not include hyphens
	pn := "183-849-2832"
	expectedLength := 10

	res, err := ProcessPhoneNumber(pn, true, false, false)

	assert.NoError(t, err)
	assert.Equal(t, len(res), expectedLength, "The length of the output phone number should be the same as the input phone number")

}

func TestProcessPhoneNumberPreserveLengthFalseHyphens(t *testing.T) {

	pn := "183-849-2832"

	res, err := ProcessPhoneNumber(pn, false, true, true)

	assert.NoError(t, err)
	assert.False(t, strings.Contains(res, "-"), "The output int phone number should not contain hyphens and may not be the same length as the input")

}

func TestProcessPhoneNumberPreserveLengthFalseNoHyphens(t *testing.T) {

	pn := "1838492832"

	res, err := ProcessPhoneNumber(pn, false, true, true)

	assert.NoError(t, err)
	assert.False(t, strings.Contains(res, "-"), "The output int phone number should not contain hyphens and may not be the same length as the input")

}

func TestProcessPhoneNumberPreserveLengthFalseIncludeHyphensTrue(t *testing.T) {

	pn := "183-849-2832"
	expectedLength := 12

	res, err := ProcessPhoneNumber(pn, false, false, true)

	assert.NoError(t, err)
	assert.Equal(t, len(res), expectedLength, "The length of the output phone number should be the same as the input phone number")

}

func TestProcessPhoneNumberPreserveLengthTrueIncludeHyphensTrueError(t *testing.T) {

	pn := "183-849-2832"
	_, err := ProcessPhoneNumber(pn, true, false, true)

	assert.Error(t, err, "The include hyphens param can only be used by itself, all other params must be false")

}

func TestProcessPhoneNumberE164Format(t *testing.T) {

	pn := "+1892393573894"
	expectedLength := 14

	res, err := ProcessPhoneNumber(pn, false, true, false)

	assert.NoError(t, err)
	assert.Equal(t, ValidateE164(res), ValidateE164("+1892393573894"))
	assert.Equal(t, len(pn), expectedLength, "The length of the output phone number should be the same as the input phone number")

}

func TestPhoneNumberTransformer(t *testing.T) {
	mapping := `root = this.phonetransformer(true, false, false)`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the phone transformer")

	testVal := "6183849282"

	res, err := ex.Query(testVal)
	assert.NoError(t, err)

	assert.Len(t, res.(string), len(testVal), "Generated phone number must be the same length as the input phone number")
}

func ValidateE164(p string) bool {

	if len(p) >= 10 && len(p) <= 15 && strings.Contains(p, "+") {
		return true
	}
	return false
}
