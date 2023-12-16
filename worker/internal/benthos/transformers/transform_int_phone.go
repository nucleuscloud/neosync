package transformers

import (
	"strconv"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(bloblang.NewBoolParam("preserve_length"))

	err := bloblang.RegisterFunctionV2("transform_int_phone", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		valuePtr, err := args.GetOptionalInt64("value")
		if err != nil {
			return nil, err
		}

		var value int64
		if valuePtr != nil {
			value = *valuePtr
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
func TransformIntPhoneNumber(number int64, preserveLength bool) (*int64, error) {

	if number == 0 {
		return nil, nil
	}

	if preserveLength {

		res, err := GenerateIntPhoneNumberPreserveLength(number)
		if err != nil {
			return nil, err
		}
		return &res, err

	} else {

		res, err := GenerateRandomIntPhoneNumber()
		if err != nil {
			return nil, err
		}

		return &res, err

	}

}

func GenerateIntPhoneNumberPreserveLength(number int64) (int64, error) {

	numStr := strconv.FormatInt(number, 10)

	length := int64(len(numStr))

	val, err := transformer_utils.GenerateRandomInt64WithInclusiveBounds(length, length)
	if err != nil {
		return 0, err
	}

	return int64(val), err

}
