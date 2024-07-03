package transformers

import (
	"errors"
	"math/rand"
	"time"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
)

// +neosyncTransformerBuilder:generate:generateBool

func init() {
	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewInt64Param("seed").Default(time.Now().UnixNano()))

	err := bloblang.RegisterFunctionV2("generate_bool", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		seed, err := args.GetInt64("seed")
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
