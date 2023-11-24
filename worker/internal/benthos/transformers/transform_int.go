package transformers

import (
	"errors"
	"fmt"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

const defaultIntLength = 4

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewInt64Param("value")).
		Param(bloblang.NewBoolParam("preserve_length"))

	err := bloblang.RegisterFunctionV2("transform_int", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		value, err := args.GetInt64("value")
		if err != nil {
			return nil, err
		}

		preserveLength, err := args.GetBool("preserve_length")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {
			res, err := TransformInt(value, preserveLength)
			return res, err
		}, nil
	})

	if err != nil {
		panic(err)
	}

}

func TransformInt(value int64, preserveLength bool) (int64, error) {

	var returnValue int64

	if transformer_utils.GetIntLength(value) > 10 {
		return 0, errors.New("the length of the input integer cannot be greater than 18 digits")
	}

	if preserveLength {

		val, err := transformer_utils.GenerateRandomInt(int(transformer_utils.GetIntLength(value)))

		if err != nil {
			return 0, fmt.Errorf("unable to generate a random string with length")
		}

		returnValue = int64(val)

	} else {

		val, err := transformer_utils.GenerateRandomInt(defaultIntLength)

		if err != nil {
			return 0, fmt.Errorf("unable to generate a random string with length")
		}

		returnValue = int64(val)

	}

	return returnValue, nil
}
