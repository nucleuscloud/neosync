package transformers

import (
	"strconv"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewInt64Param("value")).
		Param(bloblang.NewBoolParam("preserve_length"))

	// register the plugin
	err := bloblang.RegisterFunctionV2("transform_int64_phone", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		value, err := args.GetInt64("value")
		if err != nil {
			return nil, err
		}

		preserveLength, err := args.GetBool("preserve_length")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {
			res, err := TransformIntPhoneNumber(value, preserveLength)
			return res, err
		}, nil
	})

	if err != nil {
		panic(err)
	}

}

// generates a random phone number and returns it as an int64
func TransformIntPhoneNumber(number int64, preserveLength bool) (int64, error) {

	if preserveLength {

		res, err := GenerateIntPhoneNumberPreserveLength(number)
		if err != nil {
			return 0, err
		}
		return res, err

	} else {

		res, err := GenerateRandomIntPhoneNumber()
		if err != nil {
			return 0, err
		}

		return res, err

	}

}

func GenerateIntPhoneNumberPreserveLength(number int64) (int64, error) {

	numStr := strconv.FormatInt(number, 10)

	val, err := transformer_utils.GenerateRandomInt(len(numStr))
	if err != nil {
		return 0, err
	}

	return int64(val), err

}
