package neosync_transformers

import (
	"testing"
	"unicode"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestProcessRandomString(t *testing.T) {

	tests := []struct {
		s              string
		preserveLength bool
		strLength      int64
		strCase        mgmtv1alpha1.RandomString_StringCase
		expectedLength int
	}{
		{"hellothere", false, 0, mgmtv1alpha1.RandomString_STRING_CASE_LOWER, 0},   // check base string generation
		{"HELLOTHERE", false, 6, mgmtv1alpha1.RandomString_STRING_CASE_TITLE, 6},   // check string generation with a given length
		{"testtesttest", true, 0, mgmtv1alpha1.RandomString_STRING_CASE_LOWER, 12}, // check preserveLength of input string
		{"TESTEST", false, 0, mgmtv1alpha1.RandomString_STRING_CASE_UPPER, 6},      // check preserveLength of input string
	}

	for _, tt := range tests {
		res, err := ProcessRandomString(tt.s, tt.preserveLength, tt.strLength, tt.strCase)

		assert.NoError(t, err)

		if tt.preserveLength {
			assert.Equal(t, len(res), tt.expectedLength)
		}

		if !tt.preserveLength && tt.strLength == 0 {
			assert.Equal(t, len(res), 10)
		}

		if !tt.preserveLength && tt.strLength > 0 {
			assert.Equal(t, len(res), 6)
		}

		if tt.strCase == mgmtv1alpha1.RandomString_STRING_CASE_UPPER {
			assert.True(t, isAllCapitalized(res), isAllCapitalized(tt.s))
		}

		if tt.strCase == mgmtv1alpha1.RandomString_STRING_CASE_TITLE {
			assert.True(t, isTitleCased(res), isTitleCased(tt.s))
		}

		if tt.strCase == mgmtv1alpha1.RandomString_STRING_CASE_LOWER {
			assert.True(t, isAllLower(res), isAllLower(tt.s))
		}
	}

}

func TestRandomStringTransformer(t *testing.T) {
	mapping := `root = this.randomstringtransformer(true, 6, "UPPER")`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the random string transformer")

	testVal := "testte"

	res, err := ex.Query(testVal)
	assert.NoError(t, err)

	assert.Equal(t, len(testVal), len(res.(string)), "Generated string must be the same length as the input string")
	assert.IsType(t, res, testVal)
}

func isTitleCased(s string) bool {
	if s == "" {
		// An empty string is not capitalized.
		return false
	}

	// Compare the first character with its uppercase version
	return rune(s[0]) == unicode.ToUpper(rune(s[0]))
}

func isAllCapitalized(s string) bool {
	for _, char := range s {
		if unicode.IsDigit(char) {
			// Skip digits (numbers)
			continue
		}
		if !unicode.IsUpper(char) {
			return false
		}
	}
	return true
}

func isAllLower(s string) bool {
	for _, char := range s {
		if unicode.IsUpper(char) {
			return false
		}
	}
	return true
}
