package transformers

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

var (
	letterList      = "abcdefghijklmnopqrstuvwxyz"
	numberList      = "0123456789"
	specialCharList = "!@#$%^&*()-+=_ []{}|\\;\"<>,./?"
)

func init() {
	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewAnyParam("value").Optional()).Param(bloblang.NewStringParam("user_provided_regex").Optional())

	err := bloblang.RegisterFunctionV2("transform_character_scramble", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		valuePtr, err := args.GetOptionalString("value")
		if err != nil {
			return nil, err
		}

		var value string
		if valuePtr != nil {
			value = *valuePtr
		}

		regexPtr, err := args.GetOptionalString("user_provided_regex")
		if err != nil {
			return nil, err
		}

		var regex string
		if regexPtr != nil {
			regex = *regexPtr
		}

		return func() (any, error) {
			res, err := TransformCharacterScramble(value, regex)
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

func TransformCharacterScramble(value, regex string) (*string, error) {
	if value == "" {
		return nil, nil
	}

	if regex != "" {
		reg, err := regexp.Compile(regex)
		if err != nil {
			return nil, err
		}

		// finds all matches in a string
		matches := reg.FindAllStringIndex(value, -1)
		transformedString := value

		// if no matches are found just scramble the entire string
		if matches == nil {
			transformedString := strings.Map(ScrambleChar, value)
			return &transformedString, nil
		}

		// match is a [][]int with the inner []int being the start and end index values of the match
		for _, match := range matches {
			start, end := match[0], match[1]
			// run the scrambler for the substring
			matchTransformed := strings.Map(ScrambleChar, value[start:end])
			// replace the original substring with its transformed version
			transformedString = transformedString[:start] + matchTransformed + transformedString[end:]
		}

		return &transformedString, nil
	} else {
		transformedString := strings.Map(ScrambleChar, value)
		return &transformedString, nil
	}
}

func ScrambleChar(r rune) rune {
	if unicode.IsSpace(r) {
		return r
	} else if unicode.IsLetter(r) {
		randStringListInd, err := transformer_utils.GenerateRandomInt64InValueRange(0, 25)
		if err != nil {
			return r
		}
		sub := rune(letterList[randStringListInd])
		if unicode.IsUpper(r) {
			return unicode.ToUpper(sub)
		}
		return sub
	} else if unicode.IsDigit(r) {
		randNumberListInd, err := transformer_utils.GenerateRandomInt64InValueRange(0, 9)
		if err != nil {
			return r
		}

		return rune(numberList[randNumberListInd])
	} else if transformer_utils.IsAllowedSpecialChar(r) {
		randInd, err := transformer_utils.GenerateRandomInt64InValueRange(0, 28)
		if err != nil {
			return r
		}
		return rune(specialCharList[randInd])
	}

	return r
}
