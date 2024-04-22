package transformers

import (
	"fmt"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

func init() {
	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(bloblang.NewBoolParam("preserve_length")).
		Param(bloblang.NewInt64Param("min_length")).
		Param(bloblang.NewInt64Param("max_length"))

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
