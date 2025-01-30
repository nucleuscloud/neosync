package transformers

import (
	"fmt"
	"regexp"
	"testing"
	"time"
	"unicode"

	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var helloWorldRegex = "ell"
var numberRegex = "1323"
var helloTest = "helloelloelo"

func Test_ScrambleCharacter(t *testing.T) {
	testStringValue := "h"

	testRune := rune(testStringValue[0])

	res := scrambleChar(rng.New(time.Now().UnixNano()), testRune)

	assert.IsType(t, "", string(res))
	assert.Equal(t, len(testStringValue), len(string(res)), "The output string should be as long as the input string")
	assert.True(t, transformer_utils.IsValidChar(string(res)), "The output string should contain valid characters")
}

func Test_TransformCharacterSubstitutionLetters(t *testing.T) {
	testStringValue := "he11o world"

	res, err := transformCharacterScramble(rng.New(time.Now().UnixNano()), &testStringValue, "e11")

	assert.NoError(t, err)
	assert.IsType(t, "", *res)
	assert.Equal(t, len(testStringValue), len(*res), "The output string should be as long as the input string")
	assert.True(t, transformer_utils.IsValidChar(*res), "The output string should contain valid characters")
}

func Test_TransformCharacterSubstitutionCapitalizationLetters(t *testing.T) {
	testStringValue := "Hello"

	res, err := transformCharacterScramble(rng.New(time.Now().UnixNano()), &testStringValue, helloWorldRegex)

	assert.NoError(t, err)
	assert.NotNil(t, res, "Result should not be nil")
	assert.IsType(t, "", *res)
	//nolint
	assert.True(t, unicode.IsUpper([]rune(*res)[0]), "The first character of the output string should be uppercase")
	assert.Equal(t, len(testStringValue), len(*res), "The output string should be as long as the input string")
	assert.True(t, transformer_utils.IsValidChar(*res), "The output string should contain valid characters")
}

func Test_TransformCharacterSubstitutionNumbers(t *testing.T) {
	testStringValue := "41323421"

	res, err := transformCharacterScramble(rng.New(time.Now().UnixNano()), &testStringValue, numberRegex)

	assert.NoError(t, err)
	assert.IsType(t, "", *res)
	assert.Equal(t, len(testStringValue), len(*res), "The output string should be as long as the input string")
	assert.True(t, transformer_utils.IsValidChar(*res), "The output string should contain valid characters")
}

func Test_TransformCharacterSubstitutionLettersNumbers(t *testing.T) {
	testStringValue := "hello wor23r2ld 221"

	res, err := transformCharacterScramble(rng.New(time.Now().UnixNano()), &testStringValue, helloWorldRegex)

	assert.NoError(t, err)
	assert.IsType(t, "", *res)
	assert.Equal(t, len(testStringValue), len(*res), "The output string should be as long as the input string")
	assert.True(t, transformer_utils.IsValidChar(*res), "The output string should contain valid characters")
}

func Test_TransformCharacterSubstitutionLettersNumbersCharacters(t *testing.T) {
	testStringValue := "h#*(&lo wor23r2ld 221"

	res, err := transformCharacterScramble(rng.New(time.Now().UnixNano()), &testStringValue, `#\*\(&`)

	assert.NoError(t, err)
	assert.IsType(t, "", *res)
	assert.Equal(t, len(testStringValue), len(*res), "The output string should be as long as the input string")
	assert.True(t, transformer_utils.IsValidChar(*res), "The output string should contain valid characters")
}

func Test_TransformCharacterSubstitutionLettersMultipleMatches(t *testing.T) {
	// should match the first two sections and not that last i.e. h_ello_ello_elo
	res, err := transformCharacterScramble(rng.New(time.Now().UnixNano()), &helloTest, `ello`)

	assert.NoError(t, err)
	assert.IsType(t, "", *res)
	assert.Equal(t, len(helloTest), len(*res), "The output string should be as long as the input string")
	assert.True(t, transformer_utils.IsValidChar(*res), "The output string should contain valid characters")
	assert.Equal(t, helloTest[:1], (*res)[:1], "The first letter should be the same")
	assert.Equal(t, helloTest[9:], (*res)[9:], "The last three letters should be the same")
}

func Test_TransformCharacterSubstitutionLettersNoMatches(t *testing.T) {
	// should match the first two sections and not that last i.e. h_ello_ello_elo
	res, err := transformCharacterScramble(rng.New(time.Now().UnixNano()), &helloTest, `123`)

	assert.NoError(t, err)
	assert.IsType(t, "", *res)
	assert.Equal(t, len(helloTest), len(*res), "The output string should be as long as the input string")
	assert.True(t, transformer_utils.IsValidChar(*res), "The output string should contain valid characters")
	assert.False(t, helloTest == *res, "The first letter should be the same")
}

func Test_TransformCharacterSubstitutionLettersNilregex(t *testing.T) {
	// should match the first two sections and not that last i.e. h_ello_ello_elo
	res, err := transformCharacterScramble(rng.New(time.Now().UnixNano()), &testStringValue, ``)

	assert.NoError(t, err)
	assert.IsType(t, "", *res)
	assert.Equal(t, len(testStringValue), len(*res), "The output string should be as long as the input string")
	assert.True(t, transformer_utils.IsValidChar(*res), "The output string should contain valid characters")
	assert.False(t, testStringValue == *res, "The first letter should be the same")
}

func Test_TransformCharacterSubstitutionLettersMatchNumbers(t *testing.T) {
	// should match all numbers
	testStringValue := "MED-133-R123"
	complexRegex := `\d+`

	res, err := transformCharacterScramble(rng.New(time.Now().UnixNano()), &testStringValue, complexRegex)

	assert.NoError(t, err)
	assert.IsType(t, "", *res)
	assert.Equal(t, len(testStringValue), len(*res), "The output string should be as long as the input string")
	assert.True(t, transformer_utils.IsValidChar(*res), "The output string should contain valid characters")
	assert.Equal(t, testStringValue[:4], (*res)[:4], "The first letter should be the same")

	numberRegex := regexp.MustCompile(complexRegex)
	matchesOriginal := numberRegex.FindAllString(testStringValue, -1)
	matchesTransformed := numberRegex.FindAllString(*res, -1)

	// Aasert that numbers are still numbers and have the same count
	assert.Equal(t, len(matchesOriginal), len(matchesTransformed), "The number of numeric characters should be the same")
	for _, match := range matchesTransformed {
		for _, char := range match {
			assert.True(t, unicode.IsDigit(char), "Each character in the numeric matches should still be a digit")
		}
	}
}

func Test_TransformCharacterSubstitutionLettersSemiComplexRegex(t *testing.T) {
	// should match the first everything between the MED and 123)
	testStringValue := "MED-133-L123"
	complexRegex := `-(.+?)-`

	res, err := transformCharacterScramble(rng.New(time.Now().UnixNano()), &testStringValue, complexRegex)

	assert.NoError(t, err)
	assert.IsType(t, "", *res)
	assert.Equal(t, len(testStringValue), len(*res), "The output string should be as long as the input string")
	assert.True(t, transformer_utils.IsValidChar(*res), "The output string should contain valid characters")
	assert.Equal(t, testStringValue[:3], (*res)[:3], "The first letter should be the same")
	assert.Equal(t, testStringValue[10:], (*res)[10:], "The last three letters should be the same")
}

func Test_TransformCharacterSubstitutionLettersComplexRegex(t *testing.T) {
	// should match the first everything between the MED and 123)
	testStringValue := "MED-133-A123"
	complexRegex := `-(.+?)-`

	res, err := transformCharacterScramble(rng.New(time.Now().UnixNano()), &testStringValue, complexRegex)

	assert.NoError(t, err)
	assert.IsType(t, "", *res)
	assert.Equal(t, len(testStringValue), len(*res), "The output string should be as long as the input string")
	assert.True(t, transformer_utils.IsValidChar(*res), "The output string should contain valid characters")
	assert.Equal(t, testStringValue[:3], (*res)[:3], "The first letter should be the same")
	assert.Equal(t, testStringValue[10:], (*res)[10:], "The last three letters should be the same")
}

func Test_TransformCharacterSubstitutionTransformer(t *testing.T) {
	testStringValue := "hello wor23r2ld 221"

	mapping := fmt.Sprintf(`root = transform_character_scramble(value:%q,user_provided_regex:%q)`, testStringValue, helloWorldRegex)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the substitution transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.NotNil(t, res, "The response shouldn't be nil.")

	resStr, ok := res.(*string)
	if !ok {
		t.Errorf("Expected *string, got %T", res)
		return
	}

	if resStr != nil {
		assert.IsType(t, "", *resStr)
		assert.Equal(t, len(testStringValue), len(*resStr), "The output string should be as long as the input string")
		assert.True(t, transformer_utils.IsValidChar(*resStr), "The output string should contain valid characters")
	} else {
		t.Error("Pointer is nil, expected a valid string pointer")
	}
}

func Test_TransformCharacterSubsitutitionRegexEmail(t *testing.T) {
	emailregex := `(gmail\.com|yahoo\.com|nucleus\.com)$`

	testEmail := "nick@gmail.com"

	res, err := transformCharacterScramble(rng.New(time.Now().UnixNano()), &testEmail, emailregex)

	assert.NoError(t, err)
	assert.IsType(t, "", *res)
	assert.Equal(t, len(testEmail), len(*res), "The output string should be as long as the input string")
	assert.True(t, transformer_utils.IsValidChar(*res), "The output string should contain valid characters")
	assert.Equal(t, testEmail[:4], (*res)[:4], "The username should be the same")
}

func Test_TransformCharacterSubstitutionTransformer_EmptyValue(t *testing.T) {
	emptyString := ""
	mapping := fmt.Sprintf(`root = transform_character_scramble(value:%q)`, emptyString)
	ex, err := bloblang.Parse(mapping)
	require.NoError(t, err, "failed to parse the character substitution transformer")

	res, err := ex.Query(nil)
	require.NoError(t, err)
	require.NotNil(t, res, "The response shouldnt be nil")

	responseStr, ok := res.(*string)
	require.True(t, ok)
	require.NotNil(t, responseStr)
	require.Equal(t, emptyString, *responseStr)
}

func Test_TransformCharacterSubstitutionTransformer_NilValue(t *testing.T) {
	ex, err := bloblang.Parse("root = transform_character_scramble()")
	require.NoError(t, err, "failed to parse the character substitution transformer")

	res, err := ex.Query(nil)
	require.NoError(t, err)
	require.Nil(t, res, "The response was not nil")
}
