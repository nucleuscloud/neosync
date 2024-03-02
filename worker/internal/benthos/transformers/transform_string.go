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

		maxLength, err := args.GetInt64("max_length")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {
			res, err := TransformString(value, preserveLength, maxLength)
			return res, err
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

// Transforms an existing string value into another string. Does not account for numbers and other characters. If you want to preserve spaces, capitalization and other characters, use the Transform_Characters transformer.
func TransformString(value string, preserveLength bool, maxLength int64) (*string, error) {
	var returnValue string

	if value == "" {
		return nil, nil
	}

	if preserveLength {
		l := int64(len(value))
		val, err := transformer_utils.GenerateRandomStringWithDefinedLength(l)

		if err != nil {
			return nil, fmt.Errorf("unable to generate a random string with preserved length: %d: %w", l, err)
		}

		returnValue = val
	} else {
		// todo: this is a bug and we need to read in the min length based on the constraints
		min := int64(3)
		val, err := transformer_utils.GenerateRandomStringWithInclusiveBounds(min, maxLength)
		if err != nil {
			return nil, fmt.Errorf("unable to transform a random string with length: [%d:%d]: %w", min, maxLength, err)
		}
		returnValue = val
	}
	return &returnValue, nil
}
