package transformers

import (
	"fmt"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

const defaultStrLength = 10

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewStringParam("value").Optional()).
		Param(bloblang.NewBoolParam("preserve_length").Optional()).
		Param(bloblang.NewInt64Param("str_length").Optional())

	err := bloblang.RegisterFunctionV2("randomstringtransformer", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		valuePtr, err := args.GetOptionalString("value")
		if err != nil {
			return nil, err
		}

		var value string
		if valuePtr != nil {
			value = *valuePtr
		}

		preserveLengthPtr, err := args.GetOptionalBool("preserve_length")
		if err != nil {
			return nil, err
		}

		var preserveLength bool
		if preserveLengthPtr != nil {
			preserveLength = *preserveLengthPtr
		}

		strLengthPtr, err := args.GetOptionalInt64("str_length")
		if err != nil {
			return nil, err
		}

		var strLength int64
		if strLengthPtr != nil {
			strLength = *strLengthPtr
		}

		if err != nil {
			return nil, fmt.Errorf("unable to convert the string case to a defined enum value")
		}

		return func() (any, error) {
			res, err := GenerateRandomString(value, preserveLength, strLength)
			return res, err
		}, nil
	})

	if err != nil {
		panic(err)
	}

}

// main transformer logic goes here
func GenerateRandomString(value string, preserveLength bool, strLength int64) (string, error) {
	var returnValue string

	if preserveLength && strLength > 0 {
		return "", fmt.Errorf("preserve length and int length params cannot both be true")
	}

	if value != "" {

		if preserveLength {

			val, err := transformer_utils.GenerateRandomStringWithLength(int64(len(value)))

			if err != nil {
				return "", fmt.Errorf("unable to generate a random string with length")
			}

			returnValue = val

		} else if strLength > 0 {

			val, err := transformer_utils.GenerateRandomStringWithLength(strLength)

			if err != nil {
				return "", fmt.Errorf("unable to generate a random string with length")
			}

			returnValue = val

		} else {

			val, err := transformer_utils.GenerateRandomStringWithLength(defaultStrLength)

			if err != nil {
				return "", fmt.Errorf("unable to generate a random string with length")
			}

			returnValue = val

		}
	} else if strLength != 0 {

		val, err := transformer_utils.GenerateRandomStringWithLength(strLength)

		if err != nil {
			return "", fmt.Errorf("unable to generate a random string with length")
		}

		returnValue = val

	} else {

		val, err := transformer_utils.GenerateRandomStringWithLength(defaultStrLength)

		if err != nil {
			return "", fmt.Errorf("unable to generate a random string with length")
		}

		returnValue = val

	}

	return returnValue, nil
}
