package neosync_transformers

import (
	"fmt"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

const defaultIntLength = 4

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewBoolParam("preserve_length")).
		Param(bloblang.NewInt64Param("int_length"))

	// register the plugin
	err := bloblang.RegisterMethodV2("randominttransformer", spec, func(args *bloblang.ParsedParams) (bloblang.Method, error) {

		preserveLength, err := args.GetBool("preserve_length")
		if err != nil {
			return nil, err
		}

		intLength, err := args.GetInt64("int_length")
		if err != nil {
			return nil, err
		}

		return bloblang.Int64Method(func(i int64) (any, error) {
			res, err := ProcessRandomInt(i, preserveLength, intLength)
			return res, err
		}), nil
	})

	if err != nil {
		panic(err)
	}

}

// main transformer logic goes here
func ProcessRandomInt(i int64, preserveLength bool, intLength int64) (int64, error) {
	var returnValue int64

	if preserveLength {

		val, err := transformer_utils.GenerateRandomInt(transformer_utils.GetIntLength(i))

		if err != nil {
			return 0, fmt.Errorf("unable to generate a random string with length")
		}

		returnValue = val

	} else if intLength > 0 {

		val, err := transformer_utils.GenerateRandomInt(intLength)

		if err != nil {
			return 0, fmt.Errorf("unable to generate a random string with length")
		}

		returnValue = val

	} else if preserveLength && intLength > 0 {

		val, err := transformer_utils.GenerateRandomInt(transformer_utils.GetIntLength(i))

		if err != nil {
			return 0, fmt.Errorf("unable to generate a random string with length")
		}

		returnValue = val

	} else {

		val, err := transformer_utils.GenerateRandomInt(defaultIntLength)

		if err != nil {
			return 0, fmt.Errorf("unable to generate a random string with length")
		}

		returnValue = val

	}

	return returnValue, nil
}
