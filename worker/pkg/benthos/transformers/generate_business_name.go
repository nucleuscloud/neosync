package transformers

import (
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	transformers_dataset "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/data-sets"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

// +neosyncTransformerBuilder:generate:generateBusinessName

func init() {
	spec := bloblang.NewPluginSpec().
		Description("Generates a random business name between 2 and 35 characters long.").
		Category("string").
		Param(bloblang.NewInt64Param("max_length").Default(100).Description("Specifies the maximum length for the generated data. This field ensures that the output does not exceed a certain number of characters.")).
		Param(bloblang.NewInt64Param("seed").Optional().Description("An optional seed value used to generate deterministic outputs."))

	err := bloblang.RegisterFunctionV2("generate_business_name", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
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
			output, err := generateRandomBusinessName(randomizer, nil, maxLength)
			if err != nil {
				return nil, fmt.Errorf("unable to run generate_business_name: %w", err)
			}
			return output, nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

func NewGenerateBusinessNameOptsFromConfig(config *mgmtv1alpha1.GenerateBusinessName, maxLength *int64) (*GenerateBusinessNameOpts, error) {
	if config == nil {
		return NewGenerateBusinessNameOpts(nil, nil)
	}
	return NewGenerateBusinessNameOpts(
		maxLength, nil,
	)
}

func (t *GenerateBusinessName) Generate(opts any) (any, error) {
	parsedOpts, ok := opts.(*GenerateBusinessNameOpts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
	}

	return generateRandomBusinessName(parsedOpts.randomizer, nil, parsedOpts.maxLength)
}

func generateRandomBusinessName(randomizer rng.Rand, minLength *int64, maxLength int64) (string, error) {
	return transformer_utils.GenerateStringFromCorpus(
		randomizer,
		transformers_dataset.BusinessNames,
		transformers_dataset.BusinessNameMap,
		transformers_dataset.BusinessNameIndices,
		minLength,
		maxLength,
		nil,
	)
}
