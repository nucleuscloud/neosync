package transformers

import (
	"errors"
	"fmt"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
)

// +neosyncTransformerBuilder:transform:transformString

func init() {
	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(bloblang.NewBoolParam("preserve_length").Default(false).Description("Whether the original length of the input data should be preserved during transformation. If set to true, the transformation logic will ensure that the output data has the same length as the input data.")).
		Param(bloblang.NewInt64Param("min_length").Default(1).Description("Specifies the minimum length of the transformed value.")).
		Param(bloblang.NewInt64Param("max_length").Default(20).Description("Specifies the maximum length of the transformed value."))

	err := bloblang.RegisterFunctionV2("transform_string", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		value, err := args.GetOptionalString("value")
		if err != nil {
			return nil, err
		}

		preserveLength, err := args.GetBool("preserve_length")
		if err != nil {
			return nil, err
		}

		minLength, err := args.GetInt64("min_length")
		if err != nil {
			return nil, err
		}

		maxLength, err := args.GetInt64("max_length")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {
			res, err := transformString(value, preserveLength, minLength, maxLength)
			if err != nil {
				return nil, fmt.Errorf("unable to run transform_string: %w", err)
			}
			return res, nil
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

func (t *TransformString) Transform(value, opts any) (any, error) {
	parsedOpts, ok := opts.(*TransformStringOpts)
	if !ok {
		return nil, errors.New("invalid parse opts")
	}

	valueStr, ok := value.(string)
	if !ok {
		return nil, errors.New("value is not a string")
	}

	return transformString(&valueStr, parsedOpts.preserveLength, parsedOpts.minLength, parsedOpts.maxLength)
}

// Transforms an existing string value into another string. Does not account for numbers and other characters. If you want to preserve spaces, capitalization and other characters, use the Transform_Characters transformer.
func transformString(value *string, preserveLength bool, minLength, maxLength int64) (*string, error) {
	if value == nil {
		return nil, nil
	}

	minL := minLength
	maxL := maxLength

	if preserveLength {
		valueLength := int64(len(*value))
		if valueLength == 0 {
			return value, nil
		}

		minL = valueLength
		maxL = valueLength
	}
	val, err := transformer_utils.GenerateRandomStringWithInclusiveBounds(minL, maxL)
	if err != nil {
		return nil, fmt.Errorf("unable to transform a random string with length: [%d:%d]: %w", minL, maxL, err)
	}
	return &val, nil
}
