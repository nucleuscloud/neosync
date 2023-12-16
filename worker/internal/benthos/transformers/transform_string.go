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
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(bloblang.NewBoolParam("preserve_length"))

	err := bloblang.RegisterFunctionV2("transform_string", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

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

		return func() (any, error) {
			res, err := TransformString(value, preserveLength)
			return res, err
		}, nil
	})

	if err != nil {
		panic(err)
	}

}

// main transformer logic goes here
func TransformString(value string, preserveLength bool) (*string, error) {

	var returnValue string

	if value == "" {
		return nil, nil
	}

	if preserveLength {

		l := int64(len(value))
		val, err := transformer_utils.GenerateRandomString(l, l)

		if err != nil {
			return nil, fmt.Errorf("unable to generate a random string with length")
		}

		returnValue = val

	} else {
		min := int64(3)
		max := int64(12)
		val, err := transformer_utils.GenerateRandomString(min, max)

		if err != nil {
			return nil, fmt.Errorf("unable to generate a random string with length")
		}

		returnValue = val

	}

	return &returnValue, nil
}
