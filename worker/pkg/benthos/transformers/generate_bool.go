package transformers

import (
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
)

// +neosyncTransformerBuilder:generate:generateBool

func init() {
	spec := bloblang.NewPluginSpec().
		Description("Generates a random boolean value.").
		Category("boolean").
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
			return generateRandomBool(randomizer), nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

func NewGenerateBoolOptsFromConfig(config *mgmtv1alpha1.GenerateBool) (*GenerateBoolOpts, error) {
	return NewGenerateBoolOpts(nil)
}

func (t *GenerateBool) Generate(opts any) (any, error) {
	parsedOpts, ok := opts.(*GenerateBoolOpts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
	}

	return generateRandomBool(parsedOpts.randomizer), nil
}

func generateRandomBool(randomizer rng.Rand) bool {
	randInt := randomizer.Intn(2)
	return randInt == 1
}
