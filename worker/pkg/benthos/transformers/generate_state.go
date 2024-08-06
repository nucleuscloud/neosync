package transformers

import (
	"errors"

	transformers_dataset "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/data-sets"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

// +neosyncTransformerBuilder:generate:generateState

func init() {
	spec := bloblang.NewPluginSpec().Description("Randomly selects a US state and either returns the two character state code or the full state name.").
		Param(bloblang.NewBoolParam("generate_full_name").Default(false).Description("If true returns the full state name instead of the two character state code.")).
		Param(bloblang.NewInt64Param("seed").Optional().Description("An optional seed value used to generate deterministic outputs."))

	err := bloblang.RegisterFunctionV2("generate_state", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		generateFullName, err := args.GetBool("generate_full_name")
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
			return generateRandomState(randomizer, generateFullName), nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

func (t *GenerateState) Generate(opts any) (any, error) {
	parsedOpts, ok := opts.(*GenerateStateOpts)
	if !ok {
		return nil, errors.New("invalid parse opts")
	}
	return generateRandomState(parsedOpts.randomizer, parsedOpts.generateFullName), nil
}

/*
Generates a randomly selected state that exists in the United States.

By default, it returns the 2-letter state code i.e. California will return CA. However, this is configurable using the Generate Full Name parameter which, when set to true, will return the full name of the state starting with a capitalized letter.
*/
func generateRandomState(randomizer rng.Rand, generateFullName bool) string {
	randomIndex := randomizer.Intn(len(transformers_dataset.States))
	if generateFullName {
		return transformers_dataset.States[randomIndex].FullName
	}
	return transformers_dataset.States[randomIndex].Code
}
