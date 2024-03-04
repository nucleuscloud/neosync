package transformers

import (
	"fmt"

	"github.com/benthosdev/benthos/v4/public/bloblang"
)

func init() {
	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(bloblang.NewBoolParam("preserve_length")).
		Param(bloblang.NewInt64Param("max_length"))

	err := bloblang.RegisterFunctionV2("transform_phone_number", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
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
			res, err := TransformPhoneNumber(value, preserveLength, maxLength)
			if err != nil {
				return nil, fmt.Errorf("unable to run transform_phone_number: %w", err)
			}
			return res, nil
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

// Generates a random phone number and returns it as a string
func TransformPhoneNumber(value string, preserveLength bool, maxLength int64) (*string, error) {
	if value == "" {
		return nil, nil
	}

	if preserveLength {
		val, err := GenerateStringPhoneNumber(int64(len(value)), int64(len(value)), maxLength)
		if err != nil {
			return nil, err
		}

		return &val, nil
	} else {
		val, err := GenerateStringPhoneNumber(int64(9), int64(15), maxLength)
		if err != nil {
			return nil, err
		}

		return &val, nil
	}
}
