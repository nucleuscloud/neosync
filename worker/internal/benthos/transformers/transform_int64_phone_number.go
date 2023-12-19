package transformers

import (
	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformers_dataset "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/data-sets"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(bloblang.NewBoolParam("preserve_length"))

	err := bloblang.RegisterFunctionV2("transform_int64_phone_number", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

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
			res, err := TransformInt64PhoneNumber(value, preserveLength)
			return res, err
		}, nil
	})

	if err != nil {
		panic(err)
	}

}

// generates a random phone number and returns it as an int64
func TransformInt64PhoneNumber(number int64, preserveLength bool) (*int64, error) {

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

		res, err := GenerateRandomInt64PhoneNumber()
		if err != nil {
			return nil, err
		}

		return &res, err

	}

}

func GenerateIntPhoneNumberPreserveLength(number int64) (int64, error) {

	ac := transformers_dataset.AreaCodes

	// get a random area code from the areacodes data set
	randAreaCode, err := transformer_utils.GetRandomValueFromSlice[int64](ac)
	if err != nil {
		return 0, err
	}

	pn, err := transformer_utils.GenerateRandomInt64FixedLength(transformer_utils.GetInt64Length(number) - 3)
	if err != nil {
		return 0, err
	}

	return randAreaCode*1e7 + pn, nil

}
