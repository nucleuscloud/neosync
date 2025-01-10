package transformers

import (
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	transformers_dataset "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/data-sets"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

// +neosyncTransformerBuilder:generate:generateStreetAddress

type Address struct {
	Address1 string `json:"address1"`
	Address2 string `json:"address2"`
	City     string `json:"city"`
	State    string `json:"state"`
	Zipcode  string `json:"zipcode"`
}

func init() {
	spec := bloblang.NewPluginSpec().
		Description("Randomly generates a street address.").
		Param(bloblang.NewInt64Param("max_length").Default(100).Description("Specifies the maximum length for the generated data. This field ensures that the output does not exceed a certain number of characters.")).
		Param(bloblang.NewInt64Param("seed").Optional().Description("An optional seed value used to generate deterministic outputs."))

	err := bloblang.RegisterFunctionV2("generate_street_address", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
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
			res, err := generateRandomStreetAddress(randomizer, maxLength)
			if err != nil {
				return nil, fmt.Errorf("unable to run generate_street_address: %w", err)
			}
			return res, nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

func NewGenerateStreetAddressOptsFromConfig(config *mgmtv1alpha1.GenerateStreetAddress, maxLength *int64) (*GenerateStreetAddressOpts, error) {
	if config == nil {
		return NewGenerateStreetAddressOpts(nil, nil)
	}
	return NewGenerateStreetAddressOpts(
		maxLength, nil,
	)
}

func (t *GenerateStreetAddress) Generate(opts any) (any, error) {
	parsedOpts, ok := opts.(*GenerateStreetAddressOpts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
	}

	return generateRandomStreetAddress(parsedOpts.randomizer, parsedOpts.maxLength)
}

/* Generates a random street address in the United States in the format <house_number> <street name> <street ending>*/
func generateRandomStreetAddress(randomizer rng.Rand, maxLength int64) (string, error) {
	output, err := transformer_utils.GenerateStringFromCorpus(
		randomizer,
		transformers_dataset.Address_Address1s,
		transformers_dataset.Address_Address1Map,
		transformers_dataset.Address_Address1Indices,
		nil,
		maxLength,
		nil,
	)
	if err != nil {
		return transformer_utils.GenerateRandomStringWithInclusiveBounds(randomizer, 1, maxLength)
	}
	return output, nil
}
