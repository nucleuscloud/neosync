package neosync_benthos_transformers_methods

import (
	"fmt"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

const defaultStrLength = 10

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewBoolParam("preserve_length")).
		Param(bloblang.NewInt64Param("str_length"))

	// register the plugin
	err := bloblang.RegisterMethodV2("randomstringtransformer", spec, func(args *bloblang.ParsedParams) (bloblang.Method, error) {

		preserveLength, err := args.GetBool("preserve_length")
		if err != nil {
			return nil, err
		}

		strLength, err := args.GetInt64("str_length")
		if err != nil {
			return nil, err
		}

		if err != nil {
			return nil, fmt.Errorf("unable to convert the string case to a defined enum value")
		}

		return bloblang.StringMethod(func(s string) (any, error) {
			res, err := GenerateRandomString(s, preserveLength, strLength)
			return res, err
		}), nil
	})

	if err != nil {
		panic(err)
	}

}

// main transformer logic goes here
func GenerateRandomString(s string, preserveLength bool, strLength int64) (string, error) {
	var returnValue string

	if preserveLength && strLength > 0 {
		return "", fmt.Errorf("preserve length and int length params cannot both be true")
	}

	if preserveLength {

		val, err := transformer_utils.GenerateRandomStringWithLength(int64(len(s)))

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

	return returnValue, nil
}
