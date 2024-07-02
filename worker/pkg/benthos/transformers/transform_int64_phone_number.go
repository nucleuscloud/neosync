package transformers

import (
	"errors"
	"fmt"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformers_dataset "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/data-sets"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
)

// +neosyncTransformerBuilder:transform:transformInt64PhoneNumber

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
			res, err := transformInt64PhoneNumber(value, preserveLength)
			if err != nil {
				return nil, fmt.Errorf("unable to run transform_int64_phone_number: %w", err)
			}
			return res, nil
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

func (t *TransformInt64PhoneNumber) Transform(value, opts any) (any, error) {
	parsedOpts, ok := opts.(*TransformInt64PhoneNumberOpts)
	if !ok {
		return nil, errors.New("invalid parse opts")
	}

	valueInt, ok := value.(int64)
	if !ok {
		return nil, errors.New("value is not an int64")
	}

	return transformInt64PhoneNumber(valueInt, parsedOpts.preserveLength)
}

// generates a random phone number and returns it as an int64
func transformInt64PhoneNumber(number int64, preserveLength bool) (*int64, error) {
	if number == 0 {
		return nil, nil
	}

	if preserveLength {
		res, err := generateIntPhoneNumberPreserveLength(number)
		if err != nil {
			return nil, err
		}
		return &res, nil
	} else {
		res, err := generateRandomInt64PhoneNumber()
		if err != nil {
			return nil, err
		}
		return &res, nil
	}
}

func generateIntPhoneNumberPreserveLength(number int64) (int64, error) {
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
