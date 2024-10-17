package transformers

import (
	"errors"
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

// +neosyncTransformerBuilder:transform:transformFirstName

func init() {
	spec := bloblang.NewPluginSpec().
		Description("Transforms an existing first name").
		Param(bloblang.NewInt64Param("max_length").Default(100).Description("Specifies the maximum length for the transformed data. This field ensures that the output does not exceed a certain number of characters.")).
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(bloblang.NewBoolParam("preserve_length").Default(false).Description("Whether the original length of the input data should be preserved during transformation. If set to true, the transformation logic will ensure that the output data has the same length as the input data.")).
		Param(bloblang.NewInt64Param("seed").Optional().Description("An optional seed value used for generating deterministic transformations."))

	err := bloblang.RegisterFunctionV2("transform_first_name", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
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
			res, err := transformFirstName(randomizer, value, preserveLength, maxLength)
			if err != nil {
				return nil, fmt.Errorf("unable to run transform_first_name: %w", err)
			}
			return res, nil
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

func NewTransformFirstNameOptsFromConfig(config *mgmtv1alpha1.TransformFirstName, maxLength *int64) (*TransformFirstNameOpts, error) {
	if config == nil {
		return NewTransformFirstNameOpts(nil, nil, nil)
	}
	return NewTransformFirstNameOpts(
		maxLength,
		config.PreserveLength,
		nil,
	)
}

func (t *TransformFirstName) Transform(value, opts any) (any, error) {
	parsedOpts, ok := opts.(*TransformFirstNameOpts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
	}

	valueStr, ok := value.(string)
	if !ok {
		return nil, errors.New("value is not a string")
	}

	return transformFirstName(parsedOpts.randomizer, valueStr, parsedOpts.preserveLength, parsedOpts.maxLength)
}

// Generates a random first name which can be of either random length or as long as the input name
func transformFirstName(randomizer rng.Rand, value string, preserveLength bool, maxLength int64) (*string, error) {
	if value == "" {
		return &value, nil
	}

	maxValue := maxLength

	// unable to generate a random name of this fixed size
	// we may want to change this to just use the below algorithm and pad so that it is more unique
	// as with this algorithm, it will only ever use values from the underlying map that are that specific size
	if preserveLength {
		maxValue = int64(len(value))
		output, err := generateRandomFirstName(randomizer, &maxValue, maxValue)
		if err == nil {
			return &output, nil
		}
	}

	output, err := generateRandomFirstName(randomizer, nil, maxValue)
	if err != nil {
		return nil, err
	}

	// pad the string so that we can get the correct value
	if preserveLength && int64(len(output)) != maxValue {
		output += transformer_utils.GetRandomCharacterString(randomizer, maxValue-int64(len(output)))
	}
	return &output, nil
}
