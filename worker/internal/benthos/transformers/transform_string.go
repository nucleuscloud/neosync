package transformers

import (
	"fmt"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

func init() {
	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(bloblang.NewBoolParam("preserve_length")).Param(bloblang.NewInt64Param("max_length"))

	err := bloblang.RegisterFunctionV2("transform_string", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
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

		minLength, err := args.GetInt64("min_length")
		if err != nil {
			return nil, err
		}

		maxLength, err := args.GetInt64("max_length")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {
			res, err := TransformString(value, preserveLength, minLength, maxLength)
			return res, err
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

// Transforms an existing string value into another string. Does not account for numbers and other characters. If you want to preserve spaces, capitalization and other characters, use the Transform_Characters transformer.
func TransformString(value string, preserveLength bool, minLength, maxLength int64) (*string, error) {
	// todo: this is potentially a bug and we should pass in whether or not the column is nullable
	// in databases, there is a difference between a null and empty string!
	if value == "" {
		return nil, nil
	}

	if preserveLength {
		valueLength := int64(len(value))
		val, err := transformer_utils.GenerateRandomStringWithDefinedLength(valueLength)
		if err != nil {
			return nil, fmt.Errorf("unable to generate a random string with preserved length: %d: %w", valueLength, err)
		}
		return &val, nil
	}
	val, err := transformer_utils.GenerateRandomStringWithInclusiveBounds(minLength, maxLength)
	if err != nil {
		return nil, fmt.Errorf("unable to transform a random string with length: [%d:%d]: %w", minLength, maxLength, err)
	}
	return &val, nil
}
