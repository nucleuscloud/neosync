package transformers

import (
	"strings"
	"unicode"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
)

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewAnyParam("value").Optional())

	err := bloblang.RegisterFunctionV2("transform_character_substitution", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		valuePtr, err := args.GetOptionalString("value")
		if err != nil {
			return nil, err
		}

		var value string
		if valuePtr != nil {
			value = *valuePtr
		}

		return func() (any, error) {
			res, err := TransformCharacterSubstitution(value)
			return res, err
		}, nil
	})

	if err != nil {
		panic(err)
	}

}

/*
Transforms a string value with characters into an anonymized version of that string value while preserving spaces and capitalization. Letters will be replaced with letters, numbers with numbers and non-number or letter ASCII characters such as !&* with other charaacters.

For example:

Original: Hello World 123!$%
Substituted: Ifmmp Xpsme 234@%^

Note that this does not work for hex values: 0x00 -> 0x1F
*/

func TransformCharacterSubstitution(value string) (*string, error) {

	transformedString := strings.Map(substituteChar, value)

	return &transformedString, nil
}

func substituteChar(r rune) rune {
	letterMap := map[rune]rune{
		'a': 'b', 'b': 'c', 'c': 'd', 'd': 'e', 'e': 'f', 'f': 'g', 'g': 'h', 'h': 'i', 'i': 'j',
		'j': 'k', 'k': 'l', 'l': 'm', 'm': 'n', 'n': 'o', 'o': 'p', 'p': 'q', 'q': 'r', 'r': 's',
		's': 't', 't': 'u', 'u': 'v', 'v': 'w', 'w': 'x', 'x': 'y', 'y': 'z', 'z': 'a',
	}

	numberMap := map[rune]rune{
		'0': '1', '1': '2', '2': '3', '3': '4', '4': '5', '5': '6', '6': '7', '7': '8', '8': '9', '9': '0',
	}

	specialCharMap := map[rune]rune{
		'!': '@', '@': '#', '#': '$', '$': '%', '%': '^', '^': '&', '&': '*',
		'*': '(', '(': ')', ')': '-', '-': '+', '+': '=', '=': '_',
		'_': '[', '[': ']', ']': '{', '{': '}', '}': '|', '|': '\\',
		'\\': ':', ':': ';', ';': '\'', '\'': '"', '"': '<', '<': '>',
		'>': ',', ',': '.', '.': '/', '/': '?', '?': '!',
	}

	if unicode.IsLetter(r) {
		lowerR := unicode.ToLower(r)
		sub, ok := letterMap[lowerR]
		if ok {
			if unicode.IsUpper(r) {
				return unicode.ToUpper(sub)
			}
			return sub
		}
	}

	if unicode.IsDigit(r) {
		sub, ok := numberMap[r]
		if ok {
			return sub
		}
	}

	if sub, ok := specialCharMap[r]; ok {
		return sub
	}

	return r
}
