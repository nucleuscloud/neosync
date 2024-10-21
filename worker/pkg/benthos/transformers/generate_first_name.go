package transformers

import (
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	transformers_dataset "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/data-sets"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

// +neosyncTransformerBuilder:generate:generateFirstName

func init() {
	spec := bloblang.NewPluginSpec().
		Description("Generates a random first name between 2 and 12 characters long.").
		Param(bloblang.NewInt64Param("max_length").Default(100).Description("Specifies the maximum length for the generated data. This field ensures that the output does not exceed a certain number of characters.")).
		Param(bloblang.NewInt64Param("seed").Optional().Description("An optional seed value used to generate deterministic outputs."))

	err := bloblang.RegisterFunctionV2("generate_first_name", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
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
			output, err := generateRandomFirstName(randomizer, nil, maxLength)
			if err != nil {
				return nil, fmt.Errorf("unable to run generate_first_name: %w", err)
			}
			return output, nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

func NewGenerateFirstNameOptsFromConfig(config *mgmtv1alpha1.GenerateFirstName, maxLength *int64) (*GenerateFirstNameOpts, error) {
	if config == nil {
		return NewGenerateFirstNameOpts(nil, nil)
	}
	return NewGenerateFirstNameOpts(
		maxLength, nil,
	)
}

func (t *GenerateFirstName) Generate(opts any) (any, error) {
	parsedOpts, ok := opts.(*GenerateFirstNameOpts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
	}

	return generateRandomFirstName(parsedOpts.randomizer, nil, parsedOpts.maxLength)
}

func generateRandomFirstName(randomizer rng.Rand, minLength *int64, maxLength int64) (string, error) {
	return transformer_utils.GenerateStringFromCorpus(
		randomizer,
		transformers_dataset.FirstNames,
		transformers_dataset.FirstNameMap,
		transformers_dataset.FirstNameIndices,
		minLength,
		maxLength,
		nil,
	)
}
