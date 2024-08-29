package transformers

import (
	"fmt"

	transformers_dataset "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/data-sets"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

// +neosyncTransformerBuilder:generate:generateCountryCode

func init() {
	spec := bloblang.NewPluginSpec().Description("Randomly selects a Country and either returns the two character country code or the full country name.").
		Param(bloblang.NewBoolParam("generate_full_name").Default(false).Description("If true returns the full country name instead of the two character country code.")).
		Param(bloblang.NewInt64Param("seed").Optional().Description("An optional seed value used to generate deterministic outputs."))

	err := bloblang.RegisterFunctionV2("generate_country_code", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
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
			return generateRandomCountry(randomizer, generateFullName), nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

func (t *GenerateCountryCode) Generate(opts any) (any, error) {
	parsedOpts, ok := opts.(*GenerateCountryCodeOpts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
	}
	return generateRandomCountry(parsedOpts.randomizer, parsedOpts.generateFullName), nil
}

/*
Generates a randomly selected country.

By default, it returns the 2-letter country code i.e. Albania will return AL. However, this is configurable using the Generate Full Name parameter which, when set to true, will return the full name of the country starting with a capitalized letter.
*/
func generateRandomCountry(randomizer rng.Rand, generateFullName bool) string {
	randomIndex := randomizer.Intn(len(transformers_dataset.Countries))
	if generateFullName {
		return transformers_dataset.Countries[randomIndex].FullName
	}
	return transformers_dataset.Countries[randomIndex].Code
}
