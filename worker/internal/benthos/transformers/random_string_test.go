package neosync_transformers

import (
	"testing"
	"unicode"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestRandomStringPreserveLengthTrue(t *testing.T) {

	val := "hellothe"
	expectedLength := 8

	res, err := ProcessRandomString(val, true, 0, mgmtv1alpha1.RandomString_STRING_CASE_UPPER)

	assert.NoError(t, err)
	assert.Equal(t, len(res), expectedLength, "The output string should be as long as the input string")
	assert.True(t, isAllCapitalized(res), isAllCapitalized(val))

}

func TestRandomStringPreserveLengthFalse(t *testing.T) {

	val := "hello"
	expectedLength := 10

	res, err := ProcessRandomString(val, false, 0, mgmtv1alpha1.RandomString_STRING_CASE_LOWER)

	assert.NoError(t, err)
	assert.Equal(t, len(res), expectedLength, "The output string should be as long as the input string")
	assert.True(t, isAllLower(res), isAllLower(val))

}

func TestRandomStringPreserveLengthFalseStrLength(t *testing.T) {

	val := "hello"
	expectedLength := 14

	res, err := ProcessRandomString(val, false, int64(expectedLength), mgmtv1alpha1.RandomString_STRING_CASE_TITLE)

	assert.NoError(t, err)
	assert.Equal(t, len(res), expectedLength, "The output string should be as long as the input string")
	assert.True(t, isTitleCased(res), isTitleCased(val))

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
