package transformers

import (
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	transformers_dataset "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/data-sets"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

// +neosyncTransformerBuilder:generate:generateZipcode

func init() {
	spec := bloblang.NewPluginSpec().Description("Generates a randomly selected US zipcode.").
		Param(bloblang.NewInt64Param("seed").Optional().Description("An optional seed value used to generate deterministic outputs."))

	err := bloblang.RegisterFunctionV2("generate_zipcode", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
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
			val, err := generateRandomZipcode(randomizer)
			if err != nil {
				return nil, fmt.Errorf("failed to generate_zipcode: %w", err)
			}
			return val, nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

func NewGenerateZipcodeOptsFromConfig(config *mgmtv1alpha1.GenerateZipcode) (*GenerateZipcodeOpts, error) {
	return NewGenerateZipcodeOpts(nil)
}

func (t *GenerateZipcode) Generate(opts any) (any, error) {
	parsedOpts, ok := opts.(*GenerateZipcodeOpts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
	}
	return generateRandomZipcode(parsedOpts.randomizer)
}

// Generates a randomly selected zip code that exists in the United States.
func generateRandomZipcode(randomizer rng.Rand) (string, error) {
	return transformer_utils.GenerateStringFromCorpus(
		randomizer,
		transformers_dataset.Address_ZipCodes,
		transformers_dataset.Address_ZipCodeMap,
		transformers_dataset.Address_ZipCodeIndices,
		nil,
		10000,
		nil,
	)
}
