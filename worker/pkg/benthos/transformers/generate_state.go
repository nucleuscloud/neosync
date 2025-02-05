package transformers

import (
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	transformers_dataset "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/data-sets"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
)

// +neosyncTransformerBuilder:generate:generateState

func init() {
	spec := bloblang.NewPluginSpec().Description("Randomly selects a US state and by default, returns it as a 2-letter state code.").
		Category("string").
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
			val, err := generateRandomState(randomizer, generateFullName)
			if err != nil {
				return nil, fmt.Errorf("unable to run generate_state: %w", err)
			}
			return val, nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

func NewGenerateStateOptsFromConfig(config *mgmtv1alpha1.GenerateState) (*GenerateStateOpts, error) {
	if config == nil {
		return NewGenerateStateOpts(nil, nil)
	}
	return NewGenerateStateOpts(config.GenerateFullName, nil)
}

func (t *GenerateState) Generate(opts any) (any, error) {
	parsedOpts, ok := opts.(*GenerateStateOpts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
	}
	return generateRandomState(parsedOpts.randomizer, parsedOpts.generateFullName)
}

/*
Generates a randomly selected state that exists in the United States.

By default, it returns the 2-letter state code i.e. California will return CA. However, this is configurable using the Generate Full Name parameter which, when set to true, will return the full name of the state starting with a capitalized letter.
*/
func generateRandomState(randomizer rng.Rand, generateFullName bool) (string, error) {
	if generateFullName {
		return transformer_utils.GenerateStringFromCorpus(
			randomizer,
			transformers_dataset.UsStates,
			transformers_dataset.UsStateMap,
			transformers_dataset.UsStateIndices,
			nil,
			10000,
			nil,
		)
	}
	return transformer_utils.GenerateStringFromCorpus(
		randomizer,
		transformers_dataset.UsStateCodes,
		transformers_dataset.UsStateCodeMap,
		transformers_dataset.UsStateCodeIndices,
		nil,
		10000,
		nil,
	)
}
