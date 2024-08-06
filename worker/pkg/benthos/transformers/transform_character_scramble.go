package transformers

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

// +neosyncTransformerBuilder:transform:transformCharacterScramble

const (
	letterList      = "abcdefghijklmnopqrstuvwxyz"
	numberList      = "0123456789"
	specialCharList = "!@#$%^&*()-+=_ []{}|\\;\"<>,./?"
)

func init() {
	spec := bloblang.NewPluginSpec().
		Description("Transforms an existing string value by scrambling the characters while maintaining the format.").
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(bloblang.NewStringParam("user_provided_regex").Optional().Description("A custom regular expression. This regex is used to manipulate input data during the transformation process.")).
		Param(bloblang.NewInt64Param("seed").Optional().Description("An optional seed value used to generate deterministic outputs."))

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

		seedArg, err := args.GetOptionalInt64("seed")
		if err != nil {
			return nil, err
		}

		seed, err := transformer_utils.GetSeedOrDefault(seedArg)
		if err != nil {
			return nil, err
		}

		randomizer := rng.New(seed)

		return func() (any, error) {
			res, err := transformCharacterScramble(randomizer, value, regex)
			if err != nil {
				return nil, fmt.Errorf("unable to run transform_character_scramble: %w", err)
			}
			return res, nil
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

func (t *TransformCharacterScramble) Transform(value, opts any) (any, error) {
	parsedOpts, ok := opts.(*TransformCharacterScrambleOpts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
	}

	valueStr, ok := value.(string)
	if !ok {
		return nil, errors.New("value is not a string")
	}

	regex := ""
	if parsedOpts.userProvidedRegex != nil && *parsedOpts.userProvidedRegex != "" {
		regex = *parsedOpts.userProvidedRegex
	}

	return transformCharacterScramble(parsedOpts.randomizer, valueStr, regex)
}

/*
Transforms a string value with characters into an anonymized version of that string value while preserving spaces and capitalization. Letters will be replaced with letters, numbers with numbers and non-number or letter ASCII characters such as !&* with other charaacters.

For example:

Original: Hello World 123!$%
Substituted: Ifmmp Xpsme 234@%^

Note that this does not work for hex values: 0x00 -> 0x1F
*/

func transformCharacterScramble(randomizer rng.Rand, value, regex string) (*string, error) {
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
			transformedString := strings.Map(randomizedScrambleChar(randomizer), value)
			return &transformedString, nil
		}

		// match is a [][]int with the inner []int being the start and end index values of the match
		for _, match := range matches {
			start, end := match[0], match[1]
			// run the scrambler for the substring
			matchTransformed := strings.Map(randomizedScrambleChar(randomizer), value[start:end])
			// replace the original substring with its transformed version
			transformedString = transformedString[:start] + matchTransformed + transformedString[end:]
		}

		return &transformedString, nil
	} else {
		transformedString := strings.Map(randomizedScrambleChar(randomizer), value)
		return &transformedString, nil
	}
}

func randomizedScrambleChar(randomizer rng.Rand) func(r rune) rune {
	return func(r rune) rune {
		return scrambleChar(randomizer, r)
	}
}

func scrambleChar(randomizer rng.Rand, r rune) rune {
	if unicode.IsSpace(r) {
		return r
	} else if unicode.IsLetter(r) {
		randStringListInd, err := transformer_utils.GenerateRandomInt64InValueRange(randomizer, 0, 25)
		if err != nil {
			return r
		}
		sub := rune(letterList[randStringListInd])
		if unicode.IsUpper(r) {
			return unicode.ToUpper(sub)
		}
		return sub
	} else if unicode.IsDigit(r) {
		randNumberListInd, err := transformer_utils.GenerateRandomInt64InValueRange(randomizer, 0, 9)
		if err != nil {
			return r
		}

		return rune(numberList[randNumberListInd])
	} else if transformer_utils.IsAllowedSpecialChar(r) {
		randInd, err := transformer_utils.GenerateRandomInt64InValueRange(randomizer, 0, 28)
		if err != nil {
			return r
		}
		return rune(specialCharList[randInd])
	}

	return r
}
