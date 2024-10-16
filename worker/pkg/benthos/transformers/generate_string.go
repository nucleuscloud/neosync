package transformers

import (
	"fmt"

	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

// +neosyncTransformerBuilder:generate:generateString

func init() {
	spec := bloblang.NewPluginSpec().
		Description("Generates a random string of alphanumeric characters..").
		Param(bloblang.NewInt64Param("min").Default(1).Description("Specifies the minimum length for the generated string.")).
		Param(bloblang.NewInt64Param("max").Default(100).Description("Specifies the maximum length for the generated string.")).
		Param(bloblang.NewInt64Param("seed").Optional().Description("An optional seed value used to generate deterministic outputs."))

	err := bloblang.RegisterFunctionV2("generate_string", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
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
			out, err := transformer_utils.GenerateRandomStringWithInclusiveBounds(randomizer, min, max)
			if err != nil {
				return nil, fmt.Errorf("unable to run generate_string: %w", err)
			}
			return out, nil
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

func (t *GenerateString) Generate(opts any) (any, error) {
	parsedOpts, ok := opts.(*GenerateStringOpts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
	}

	return transformer_utils.GenerateRandomStringWithInclusiveBounds(parsedOpts.randomizer, parsedOpts.min, parsedOpts.max)
}
