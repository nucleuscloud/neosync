package transformers

import (
	"errors"
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

// +neosyncTransformerBuilder:transform:transformLastName

func init() {
	spec := bloblang.NewPluginSpec().
		Description("Anonymizes and transforms an existing last name.").
		Category("string").
		Param(bloblang.NewInt64Param("max_length").Default(100).Description("Specifies the maximum length for the transformed data. This field ensures that the output does not exceed a certain number of characters.")).
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(bloblang.NewBoolParam("preserve_length").Default(false).Description("Whether the original length of the input data should be preserved during transformation. If set to true, the transformation logic will ensure that the output data has the same length as the input data.")).
		Param(bloblang.NewInt64Param("seed").Optional().Description("An optional seed value used for generating deterministic transformations."))

	err := bloblang.RegisterFunctionV2("transform_last_name", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
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
			res, err := transformLastName(randomizer, value, preserveLength, maxLength)
			if err != nil {
				return nil, fmt.Errorf("unable to run transform_last_name: %w", err)
			}
			return res, nil
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

func NewTransformLastNameOptsFromConfig(config *mgmtv1alpha1.TransformLastName, maxLength *int64) (*TransformLastNameOpts, error) {
	if config == nil {
		return NewTransformLastNameOpts(nil, nil, nil)
	}
	return NewTransformLastNameOpts(
		maxLength,
		config.PreserveLength,
		nil,
	)
}

func (t *TransformLastName) Transform(value, opts any) (any, error) {
	parsedOpts, ok := opts.(*TransformLastNameOpts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
	}

	valueStr, ok := value.(string)
	if !ok {
		return nil, errors.New("value is not a string")
	}

	return transformLastName(parsedOpts.randomizer, valueStr, parsedOpts.preserveLength, parsedOpts.maxLength)
}

// Generates a random last name which can be of either random length between [2,12] characters or as long as the input name
func transformLastName(randomizer rng.Rand, name string, preserveLength bool, maxLength int64) (*string, error) {
	if name == "" {
		return nil, nil
	}

	maxValue := maxLength

	// unable to generate a random name of this fixed size
	// we may want to change this to just use the below algorithm and pad so that it is more unique
	// as with this algorithm, it will only ever use values from the underlying map that are that specific size
	if preserveLength {
		maxValue = int64(len(name))
		output, err := generateRandomLastName(randomizer, &maxValue, maxValue)
		if err == nil {
			return &output, nil
		}
	}

	output, err := generateRandomLastName(randomizer, nil, maxValue)
	if err != nil {
		return nil, err
	}

	// pad the string so that we can get the correct value
	if preserveLength && int64(len(output)) != maxValue {
		output += transformer_utils.GetRandomCharacterString(randomizer, maxValue-int64(len(output)))
	}
	return &output, nil
}
