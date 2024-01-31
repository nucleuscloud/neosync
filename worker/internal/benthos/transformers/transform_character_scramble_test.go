package transformers

import (
	"fmt"
	"testing"
	"unicode"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
	"github.com/stretchr/testify/assert"
)

func Test_ScrambleCharacter(t *testing.T) {

	testStringValue := "h"

	testRune := rune(testStringValue[0])

	res := ScrambleChar(testRune)

	assert.IsType(t, "", string(res))
	assert.Equal(t, len(testStringValue), len(string(res)), "The output string should be as long as the input string")
	assert.True(t, transformer_utils.IsValidChar(string(res)), "The output string should contain valid characters")
}

func Test_TransformCharacterSubstitutionLetters(t *testing.T) {

	testStringValue := "hello world"

	res, err := TransformCharacterScramble(testStringValue)

	assert.NoError(t, err)
	assert.IsType(t, "", *res)
	assert.Equal(t, len(testStringValue), len(*res), "The output string should be as long as the input string")
	assert.True(t, transformer_utils.IsValidChar(*res), "The output string should contain valid characters")
}

func Test_TransformCharacterSubstitutionCapitalizationLetters(t *testing.T) {
	testStringValue := "Hello"

	res, err := TransformCharacterScramble(testStringValue)

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

	res, err := TransformCharacterScramble(testStringValue)

	assert.NoError(t, err)
	assert.IsType(t, "", *res)
	assert.Equal(t, len(testStringValue), len(*res), "The output string should be as long as the input string")
	assert.True(t, transformer_utils.IsValidChar(*res), "The output string should contain valid characters")
}

func Test_TransformCharacterSubstitutionLettersNumbers(t *testing.T) {

	testStringValue := "hello wor23r2ld 221"

	res, err := TransformCharacterScramble(testStringValue)

	assert.NoError(t, err)
	assert.IsType(t, "", *res)
	assert.Equal(t, len(testStringValue), len(*res), "The output string should be as long as the input string")
	assert.True(t, transformer_utils.IsValidChar(*res), "The output string should contain valid characters")
}

func Test_TransformCharacterSubstitutionLettersNumbersCharacters(t *testing.T) {

	testStringValue := "h#*(&lo wor23r2ld 221"

	res, err := TransformCharacterScramble(testStringValue)

	assert.NoError(t, err)
	assert.IsType(t, "", *res)
	assert.Equal(t, len(testStringValue), len(*res), "The output string should be as long as the input string")
	assert.True(t, transformer_utils.IsValidChar(*res), "The output string should contain valid characters")
}

func Test_TransformCharacterSubstitutionTransformer(t *testing.T) {

	testStringValue := "h#*(&lo wor23r2ld 221"

	mapping := fmt.Sprintf(`root = transform_character_scramble(value:%q)`, testStringValue)
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

func Test_TransformCharacterSubstitutionTransformerWithEmptyValue(t *testing.T) {

	nilString := ""
	mapping := fmt.Sprintf(`root = transform_character_scramble(value:%q)`, nilString)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the character substitution transformer")

	_, err = ex.Query(nil)
	assert.NoError(t, err)
}
