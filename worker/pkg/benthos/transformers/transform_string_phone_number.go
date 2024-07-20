package transformers

import (
	"errors"
	"fmt"

	"github.com/warpstreamlabs/bento/public/bloblang"
)

// +neosyncTransformerBuilder:transform:transformStringPhoneNumber

func init() {
	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(bloblang.NewBoolParam("preserve_length").Description("Whether the original length of the input data should be preserved during transformation. If set to true, the transformation logic will ensure that the output data has the same length as the input data.")).
		Param(bloblang.NewInt64Param("max_length").Description("Specifies the maximum length for the transformed data. This field ensures that the output does not exceed a certain number of characters."))

	err := bloblang.RegisterFunctionV2("transform_phone_number", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		value, err := args.GetOptionalString("value")
		if err != nil {
			return nil, err
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
			res, err := transformPhoneNumber(value, preserveLength, maxLength)
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

func (t *TransformStringPhoneNumber) Transform(value, opts any) (any, error) {
	parsedOpts, ok := opts.(*TransformStringPhoneNumberOpts)
	if !ok {
		return nil, errors.New("invalid parse opts")
	}

	valueStr, ok := value.(string)
	if !ok {
		return nil, errors.New("value is not a string")
	}

	return transformPhoneNumber(&valueStr, parsedOpts.preserveLength, parsedOpts.maxLength)
}

// Generates a random phone number and returns it as a string
func transformPhoneNumber(value *string, preserveLength bool, maxLength int64) (*string, error) {
	if value == nil {
		return nil, nil
	}

	minL := int64(9)
	maxL := maxLength

	if preserveLength {
		valueLength := int64(len(*value))
		if valueLength == 0 {
			return value, nil
		}
		minL = valueLength
		maxL = valueLength
	}
	val, err := generateStringPhoneNumber(minL, maxL)
	if err != nil {
		return nil, fmt.Errorf("unable to transform phone number with length: [%d:%d]: %w", minL, maxL, err)
	}
	return &val, nil
}
