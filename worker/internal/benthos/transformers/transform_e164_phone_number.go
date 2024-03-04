package transformers

import (
	"fmt"
	"strings"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

func init() {
	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(bloblang.NewBoolParam("preserve_length")).
		Param(bloblang.NewInt64Param("max_length"))

	err := bloblang.RegisterFunctionV2("transform_e164_phone_number", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
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
			res, err := TransformE164PhoneNumber(value, preserveLength, maxLength)
			if err != nil {
				return nil, fmt.Errorf("unable to run transform_e164_phone_number: %w", err)
			}
			return res, nil
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

// Generates a random phone number and returns it as a string
func TransformE164PhoneNumber(phone string, preserveLength bool, maxLength int64) (*string, error) {
	var returnValue string

	if phone == "" {
		return nil, nil
	}

	if preserveLength {
		res, err := GenerateE164FormatPhoneNumberPreserveLength(phone)
		if err != nil {
			return nil, err
		}

		returnValue = res
	} else {
		min := int64(9)
		max := int64(15)

		res, err := GenerateInternationalPhoneNumber(min, max)
		if err != nil {
			return nil, err
		}
		returnValue = res
	}

	return &returnValue, nil
}

// generates a random E164 phone number and returns it as a string
func GenerateE164FormatPhoneNumberPreserveLength(number string) (string, error) {
	val := strings.Split(number, "+")

	length := int64(len(val[1]))

	vals, err := transformer_utils.GenerateRandomInt64FixedLength(length)
	if err != nil {
		return "", nil
	}
	return fmt.Sprintf("+%d", vals), nil
}
