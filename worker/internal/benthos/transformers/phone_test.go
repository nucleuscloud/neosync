package neosync_transformers

import (
	"strings"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func TestProcessPhoneNumber(t *testing.T) {

	tests := []struct {
		pn              string
		preserveLength  bool
		e164_format     bool
		include_hyphens bool
		expectedLength  int
	}{
		{"6183849282", false, false, false, 0},    // check base phone number generation
		{"618384928322", true, false, false, 12},  // checks preserve length
		{"739-892-9234", false, false, true, 0},   // checks hyphens
		{"+1892393573894", false, true, false, 0}, // checks e164 format
		{"+18923935738", false, true, false, 13},  // checks e164 format
	}

	for _, tt := range tests {
		res, err := ProcessPhoneNumber(tt.pn, tt.preserveLength, tt.e164_format, tt.include_hyphens)

		assert.NoError(t, err)

		if tt.preserveLength && !tt.e164_format && !tt.include_hyphens {
			assert.Equal(t, len(strings.ReplaceAll(res, "-", "")), tt.expectedLength)
		}

		if tt.e164_format && !tt.preserveLength && !tt.include_hyphens {
			assert.Equal(t, Validatee164(res), Validatee164("+1892393573894"))
		}

		if !tt.preserveLength && !tt.e164_format && !tt.include_hyphens {
			assert.Equal(t, len(res), len("6183849282"))
		}

		if !tt.preserveLength && !tt.e164_format && tt.include_hyphens {
			assert.Equal(t, len(res), len("618-384-9282"))
		}

		if tt.preserveLength && tt.e164_format && !tt.include_hyphens {
			assert.Equal(t, len(res), tt.expectedLength)
		}

	}

}

func Validatee164(p string) bool {

	if len(p) >= 10 && len(p) <= 15 && strings.Contains(p, "+") {
		return true
	}
	return false
}

func TestPhoneNumberTransformer(t *testing.T) {
	mapping := `root = this.phonetransformer(true, false, true)`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the phone transformer")

	testVal := "6183849282"

	res, err := ex.Query(testVal)
	assert.NoError(t, err)

	assert.Len(t, res.(string), len(testVal), "Generated phone number must be the same length as the input phone number")
}
