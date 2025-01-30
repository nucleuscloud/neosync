package transformers

import (
	"errors"
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
)

// +neosyncTransformerBuilder:transform:transformStringPhoneNumber

func init() {
	spec := bloblang.NewPluginSpec().
		Description("Anonymizes and transforms an existing phone number that is typed as a string.").
		Category("string").
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(bloblang.NewBoolParam("preserve_length").Default(false).Description("Whether the original length of the input data should be preserved during transformation. If set to true, the transformation logic will ensure that the output data has the same length as the input data.")).
		Param(bloblang.NewInt64Param("max_length").Default(100).Description("Specifies the maximum length for the transformed data. This field ensures that the output does not exceed a certain number of characters.")).
		Param(bloblang.NewInt64Param("seed").Optional().Description("An optional seed value used to generate deterministic outputs."))

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
			res, err := transformPhoneNumber(randomizer, value, preserveLength, maxLength)
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

func NewTransformStringPhoneNumberOptsFromConfig(config *mgmtv1alpha1.TransformPhoneNumber, maxLength *int64) (*TransformStringPhoneNumberOpts, error) {
	if config == nil {
		return NewTransformStringPhoneNumberOpts(nil, nil, nil)
	}
	return NewTransformStringPhoneNumberOpts(
		config.PreserveLength,
		maxLength, nil,
	)
}

func (t *TransformStringPhoneNumber) Transform(value, opts any) (any, error) {
	parsedOpts, ok := opts.(*TransformStringPhoneNumberOpts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
	}

	valueStr, ok := value.(string)
	if !ok {
		return nil, errors.New("value is not a string")
	}

	return transformPhoneNumber(parsedOpts.randomizer, &valueStr, parsedOpts.preserveLength, parsedOpts.maxLength)
}

// Generates a random phone number and returns it as a string
func transformPhoneNumber(randomizer rng.Rand, value *string, preserveLength bool, maxLength int64) (*string, error) {
	if value == nil || *value == "" {
		return value, nil
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
	val, err := generateStringPhoneNumber(randomizer, minL, maxL)
	if err != nil {
		return nil, fmt.Errorf("unable to transform phone number with length: [%d:%d]: %w", minL, maxL, err)
	}
	return &val, nil
}
