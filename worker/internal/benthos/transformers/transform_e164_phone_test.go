package transformers

import (
	"fmt"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

var testE164Phone = "+13782983927"

func Test_TransformE164NumberPreserveLengthTrue(t *testing.T) {

	res, err := TransformE164Number(testE164Phone, true)

	assert.NoError(t, err)
	assert.Equal(t, ValidateE164(res), ValidateE164(testE164Phone))
	assert.Len(t, res, len(testE164Phone), "Generated phone number must be the same length as the input phone number")
}

func Test_TransformE164NumberPreserveLengthFalse(t *testing.T) {

	res, err := TransformE164Number(testE164Phone, false)

	assert.NoError(t, err)
	assert.Equal(t, ValidateE164(res), ValidateE164(testE164Phone))
	// + 1 to account for the plus sign at the beginning
	assert.Len(t, res, defaultE164Length+1, "Generated phone number must be the same length as the input phone number")
}

func Test_GenerateE164FormatPhoneNumberPreserveLength(t *testing.T) {

	res, err := GenerateE164FormatPhoneNumberPreserveLength(testE164Phone)

	assert.NoError(t, err)
	assert.Equal(t, ValidateE164(res), ValidateE164(testE164Phone))
	// + 1 to account for the plus sign at the beginning
	assert.Len(t, res, len(testE164Phone), "Generated phone number must be the same length as the input phone number")
}

func Test_TransformE164NumberTransformer(t *testing.T) {
	mapping := fmt.Sprintf(`root = transform_e164_phone(%q, true)`, testE164Phone)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the phone transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.Equal(t, ValidateE164(res.(string)), ValidateE164(testE164Phone))
	assert.Len(t, res.(string), len(testE164Phone), "Generated phone number must be the same length as the input phone number")
}
