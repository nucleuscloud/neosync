package transformers

import (
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
		Param(bloblang.NewAnyParam("value").Optional())

	err := bloblang.RegisterFunctionV2("transform_character_scramble", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		valuePtr, err := args.GetOptionalString("value")
		if err != nil {
			return nil, err
		}

		var value string
		if valuePtr != nil {
			value = *valuePtr
		}

		return func() (any, error) {
			res, err := TransformCharacterScramble(value)
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

func TransformCharacterScramble(value string) (*string, error) {
	transformedString := strings.Map(ScrambleChar, value)

	return &transformedString, nil
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
