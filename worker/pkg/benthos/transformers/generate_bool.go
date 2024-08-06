package transformers

import (
	"errors"
	"math/rand"

	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

// +neosyncTransformerBuilder:generate:generateBool

func init() {
	spec := bloblang.NewPluginSpec().
		Description("Generates a boolean value at random.").
		Param(bloblang.NewInt64Param("seed").Optional().Description("An optional seed value used to generate deterministic outputs."))

	err := bloblang.RegisterFunctionV2("generate_bool", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
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
			return generateRandomizerBool(randomizer), nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

func (t *GenerateBool) Generate(opts any) (any, error) {
	parsedOpts, ok := opts.(*GenerateBoolOpts)
	if !ok {
		return nil, errors.New("invalid parse opts")
	}

	return generateRandomizerBool(parsedOpts.randomizer), nil
}

// Generates a random bool value and returns it as a bool type.
func generateRandomBool() bool {
	//nolint:gosec
	randInt := rand.Intn(2)
	return randInt == 1
}

func generateRandomizerBool(randomizer rng.Rand) bool {
	randInt := randomizer.Intn(2)
	return randInt == 1
}
