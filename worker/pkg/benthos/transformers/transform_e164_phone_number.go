package transformers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
)

// +neosyncTransformerBuilder:transform:transformE164PhoneNumber

func init() {
	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(bloblang.NewBoolParam("preserve_length").Description("Whether the original length of the input data should be preserved during transformation. If set to true, the transformation logic will ensure that the output data has the same length as the input data.")).
		Param(bloblang.NewInt64Param("max_length").Optional().Description("Specifies the maximum length for the transformed data. This field ensures that the output does not exceed a certain number of characters."))

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

		maxLength, err := args.GetOptionalInt64("max_length")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {
			res, err := transformE164PhoneNumber(value, preserveLength, maxLength)
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

func (t *TransformE164PhoneNumber) Transform(value, opts any) (any, error) {
	parsedOpts, ok := opts.(*TransformE164PhoneNumberOpts)
	if !ok {
		return nil, errors.New("invalid parse opts")
	}

	valueStr, ok := value.(string)
	if !ok {
		return nil, errors.New("value is not a string")
	}

	return transformE164PhoneNumber(valueStr, parsedOpts.preserveLength, parsedOpts.maxLength)
}

// Generates a random phone number and returns it as a string
func transformE164PhoneNumber(phone string, preserveLength bool, maxLength *int64) (*string, error) {
	var returnValue string

	if phone == "" {
		return nil, nil
	}

	if preserveLength {
		res, err := generateE164FormatPhoneNumberPreserveLength(phone)
		if err != nil {
			return nil, err
		}

		returnValue = res
	} else {
		minValue := int64(9)
		maxValue := int64(15)
		if maxLength != nil && *maxLength != 0 {
			maxValue = *maxLength
		}

		res, err := generateInternationalPhoneNumber(minValue, maxValue)
		if err != nil {
			return nil, err
		}
		returnValue = res
	}

	return &returnValue, nil
}

// generates a random E164 phone number and returns it as a string
func generateE164FormatPhoneNumberPreserveLength(number string) (string, error) {
	val := strings.Split(number, "+")

	length := int64(len(val[1]))

	vals, err := transformer_utils.GenerateRandomInt64FixedLength(length)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("+%d", vals), nil
}
