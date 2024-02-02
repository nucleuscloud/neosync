package transformers

import (
	"fmt"
	"strings"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewInt64Param("max_length")).
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(bloblang.NewBoolParam("preserve_length"))

	err := bloblang.RegisterFunctionV2("transform_full_name", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		valuePtr, err := args.GetOptionalString("value")
		if err != nil {
			return nil, err
		}

		var value string
		if valuePtr != nil {
			value = *valuePtr
		}

		preserveLength, err := args.GetBool("preserve_length")
		if err != nil {
			return nil, err
		}

		maxLength, err := args.GetInt64("max_length")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {
			res, err := TransformFullName(value, preserveLength, maxLength)
			return res, err
		}, nil
	})

	if err != nil {
		panic(err)
	}

}

func TransformFullName(name string, preserveLength bool, maxLength int64) (*string, error) {

	if name == "" {
		return nil, nil
	}

	fnLength := int64(len(strings.Split(name, " ")[0]))

	lnLength := int64(len(strings.Split(name, " ")[1]))

	if preserveLength {
		// assume that if pl is true than it already meets the maxCharacterLimit constraint
		fn, err := GenerateRandomFirstNameInLengthRange(fnLength, fnLength)
		if err != nil {
			return nil, err
		}

		ln, err := GenerateRandomLastNameInLengthRange(lnLength, lnLength)
		if err != nil {
			return nil, err
		}

		res := fn + " " + ln
		return &res, nil

	} else {

		res, err := GenerateRandomFullNameInLengthRange(fnLength, lnLength, minNameLength, maxLength)
		if err != nil {
			return nil, err
		}
		return &res, nil
	}

}

// Generates a random full name name with length [min, max]. If the length is greater than 12, a full name of length 12 will be returned.
func GenerateRandomFullNameInLengthRange(fnLength, lnLength int64, minLength, maxLength int64) (string, error) {

	fmt.Println("fn ln", fnLength, lnLength)

	if maxLength < 12 && maxLength >= 5 {

		// calc lengths of a first name and last name, also assumes a min maxLength of 4, since first and last names must be at least 2 letters in length, we will use the remainder for the space character
		half := maxLength / 2
		firstNameLength := half
		lastNameLength := half

		fmt.Println("max", maxLength)

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
