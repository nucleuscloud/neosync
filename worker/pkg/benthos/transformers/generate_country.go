package transformers

import (
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	transformers_dataset "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/data-sets"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
)

// +neosyncTransformerBuilder:generate:generateCountry

func init() {
	spec := bloblang.NewPluginSpec().Description("Randomly selects a country and by default, returns it as a 2-letter country code.").
		Category("string").
		Param(bloblang.NewBoolParam("generate_full_name").Default(false).Description("If true returns the full country name instead of the two character country code.")).
		Param(bloblang.NewInt64Param("seed").Optional().Description("An optional seed value used to generate deterministic outputs."))

	err := bloblang.RegisterFunctionV2("generate_country", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
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
			val, err := generateRandomCountry(randomizer, generateFullName)
			if err != nil {
				return nil, fmt.Errorf("failed to generate_country: %w", err)
			}
			return val, nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

func NewGenerateCountryOptsFromConfig(config *mgmtv1alpha1.GenerateCountry) (*GenerateCountryOpts, error) {
	if config == nil {
		return NewGenerateCountryOpts(
			nil,
			nil,
		)
	}
	return NewGenerateCountryOpts(
		config.GenerateFullName, nil,
	)
}

func (t *GenerateCountry) Generate(opts any) (any, error) {
	parsedOpts, ok := opts.(*GenerateCountryOpts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
	}
	return generateRandomCountry(parsedOpts.randomizer, parsedOpts.generateFullName)
}

/*
Generates a randomly selected country.

By default, it returns the 2-letter country code i.e. Albania will return AL. However, this is configurable using the Generate Full Name parameter which, when set to true, will return the full name of the country starting with a capitalized letter.
*/
func generateRandomCountry(randomizer rng.Rand, generateFullName bool) (string, error) {
	if generateFullName {
		return transformer_utils.GenerateStringFromCorpus(
			randomizer,
			transformers_dataset.Countrys,
			transformers_dataset.CountryMap,
			transformers_dataset.CountryIndices,
			nil,
			10000,
			nil,
		)
	}
	return transformer_utils.GenerateStringFromCorpus(
		randomizer,
		transformers_dataset.CountryCodes,
		transformers_dataset.CountryCodeMap,
		transformers_dataset.CountryCodeIndices,
		nil,
		10000,
		nil,
	)
}
