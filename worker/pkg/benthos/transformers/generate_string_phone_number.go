package transformers

import (
	"errors"
	"fmt"

	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

// +neosyncTransformerBuilder:generate:generateStringPhoneNumber

func init() {
	spec := bloblang.NewPluginSpec().
		Description("Generates a Generate phone number and returns it as a string.").
		Param(bloblang.NewInt64Param("min").Description("Specifies the minimum length for the generated phone number.")).
		Param(bloblang.NewInt64Param("max").Description("Specifies the maximum length for the generated phone number.")).
		Param(bloblang.NewInt64Param("seed").Optional().Description("An optional seed value used to generate deterministic outputs."))

	err := bloblang.RegisterFunctionV2("generate_string_phone_number", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		min, err := args.GetInt64("min")
		if err != nil {
			return nil, err
		}

		max, err := args.GetInt64("max")
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
			res, err := generateStringPhoneNumber(randomizer, min, max)
			if err != nil {
				return nil, fmt.Errorf("unable to run generate_string_phone_number: %w", err)
			}
			return res, nil
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

func (t *GenerateStringPhoneNumber) Generate(opts any) (any, error) {
	parsedOpts, ok := opts.(*GenerateStringPhoneNumberOpts)
	if !ok {
		return nil, errors.New("invalid parse opts")
	}

	return generateStringPhoneNumber(parsedOpts.randomizer, parsedOpts.min, parsedOpts.max)
}

/*  Generates a string phone number in the length interval [min, max] with the min length == 9 and the max length == 15.
 */
func generateStringPhoneNumber(randomizer rng.Rand, minValue, maxValue int64) (string, error) {
	minValue = transformer_utils.Floor(minValue, 9)
	maxValue = transformer_utils.Ceil(maxValue, 15)

	val, err := transformer_utils.GenerateRandomInt64InLengthRange(randomizer, minValue, maxValue)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d", val), nil
}
