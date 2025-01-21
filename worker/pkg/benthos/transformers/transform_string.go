package transformers

import (
	"errors"
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

// +neosyncTransformerBuilder:transform:transformString

func init() {
	spec := bloblang.NewPluginSpec().
		Description("Anonymizes and transforms an existing string value.").
		Category("string").
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(bloblang.NewBoolParam("preserve_length").Default(false).Description("Whether the original length of the input data should be preserved during transformation. If set to true, the transformation logic will ensure that the output data has the same length as the input data.")).
		Param(bloblang.NewInt64Param("min_length").Default(1).Description("Specifies the minimum length of the transformed value.")).
		Param(bloblang.NewInt64Param("max_length").Default(100).Description("Specifies the maximum length of the transformed value.")).
		Param(bloblang.NewInt64Param("seed").Optional().Description("An optional seed value used to generate deterministic outputs."))

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

		seedArg, err := args.GetOptionalInt64("seed")
		if err != nil {
			return nil, err
		}

		seed, err := transformer_utils.GetSeedOrDefault(seedArg)
		if err != nil {
			return nil, err
		}

		randomizer := rng.New(seed)

		return func() (any, error) {
			res, err := transformString(randomizer, value, preserveLength, minLength, maxLength)
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

func NewTransformStringOptsFromConfig(config *mgmtv1alpha1.TransformString, minLength, maxLength *int64) (*TransformStringOpts, error) {
	if config == nil {
		return NewTransformStringOpts(nil, nil, nil, nil)
	}
	return NewTransformStringOpts(
		config.PreserveLength,
		minLength,
		maxLength,
		nil,
	)
}

func (t *TransformString) Transform(value, opts any) (any, error) {
	parsedOpts, ok := opts.(*TransformStringOpts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
	}

	valueStr, ok := value.(string)
	if !ok {
		return nil, errors.New("value is not a string")
	}

	return transformString(parsedOpts.randomizer, &valueStr, parsedOpts.preserveLength, parsedOpts.minLength, parsedOpts.maxLength)
}

// Transforms an existing string value into another string. Does not account for numbers and other characters. If you want to preserve spaces, capitalization and other characters, use the Transform_Characters transformer.
func transformString(randomizer rng.Rand, value *string, preserveLength bool, minLength, maxLength int64) (*string, error) {
	if value == nil || *value == "" {
		return value, nil
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
	val, err := transformer_utils.GenerateRandomStringWithInclusiveBounds(randomizer, minL, maxL)
	if err != nil {
		return nil, fmt.Errorf("unable to transform a random string with length: [%d:%d]: %w", minL, maxL, err)
	}
	return &val, nil
}
