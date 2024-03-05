package transformers

import (
	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

func init() {
	spec := bloblang.NewPluginSpec().Param(bloblang.NewInt64Param("max_length"))

	err := bloblang.RegisterFunctionV2("generate_full_name", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		maxLength, err := args.GetInt64("max_length")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {
			res, err := GenerateRandomFullName(maxLength)
			return res, err
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

/* Generates a random full name */
func GenerateRandomFullName(maxLength int64) (string, error) {
	if maxLength < 12 && maxLength >= 5 {
		// calc lengths of a first name and last name, also assumes a min maxLength of 4, since first and last names must be at least 2 letters in length, we will use the remainder for the space character
		half := maxLength / 2
		firstNameLength := half
		lastNameLength := half

		fn, err := GenerateRandomFirstNameInLengthRange(minNameLength, firstNameLength)
		if err != nil {
			return "", err
		}
		// the -1 accounts for the space
		ln, err := GenerateRandomLastNameInLengthRange(minNameLength, lastNameLength-1)
		if err != nil {
			return "", err
		}

		res := fn + " " + ln
		return res, nil
	} else if maxLength < 5 {
		res, err := transformer_utils.GenerateRandomStringWithDefinedLength(maxLength)
		if err != nil {
			return "", err
		}
		return res, nil
	} else {
		// assume that if pl is true than it already meets the maxCharacterLimit constraint
		fn, err := GenerateRandomFirstNameInLengthRange(int64(3), int64(6))
		if err != nil {
			return "", err
		}

		ln, err := GenerateRandomLastNameInLengthRange(int64(3), int64(6))
		if err != nil {
			return "", err
		}

		res := fn + " " + ln
		return res, nil
	}
}
