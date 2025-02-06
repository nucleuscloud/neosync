package transformers

import (
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	transformers_dataset "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/data-sets"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
)

// +neosyncTransformerBuilder:generate:generateCity

func init() {
	spec := bloblang.NewPluginSpec().
		Description("Randomly selects a city from a list of predefined US cities.").
		Category("string").
		Param(bloblang.NewInt64Param("max_length").Default(100).Description("Specifies the maximum length for the generated data. This field ensures that the output does not exceed a certain number of characters.")).
		Param(bloblang.NewInt64Param("seed").Optional().Description("An optional seed value used to generate deterministic outputs."))

	err := bloblang.RegisterFunctionV2("generate_city", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
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
			res, err := generateRandomCity(randomizer, maxLength)
			if err != nil {
				return nil, fmt.Errorf("unable to run generate_city: %w", err)
			}
			return res, nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

func NewGenerateCityOptsFromConfig(config *mgmtv1alpha1.GenerateCity, maxLength *int64) (*GenerateCityOpts, error) {
	if config == nil {
		return NewGenerateCityOpts(
			nil,
			nil,
		)
	}
	return NewGenerateCityOpts(
		maxLength, nil,
	)
}

func (t *GenerateCity) Generate(opts any) (any, error) {
	parsedOpts, ok := opts.(*GenerateCityOpts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
	}

	return generateRandomCity(parsedOpts.randomizer, parsedOpts.maxLength)
}

// Generates a randomly selected city that exists in the United States. Accounts for the maxLength of the column and searches for a city that is shorter than the maxLength.
// If not, it randomly generates a string that len(string) == maxLength
func generateRandomCity(randomizer rng.Rand, maxLength int64) (string, error) {
	output, err := transformer_utils.GenerateStringFromCorpus(
		randomizer,
		transformers_dataset.Address_Citys,
		transformers_dataset.Address_CityMap,
		transformers_dataset.Address_CityIndices,
		nil,
		maxLength,
		nil,
	)
	if err != nil {
		return transformer_utils.GenerateRandomStringWithInclusiveBounds(randomizer, 1, maxLength)
	}
	return output, nil
}
