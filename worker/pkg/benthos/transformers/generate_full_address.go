package transformers

import (
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	transformers_dataset "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/data-sets"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

// +neosyncTransformerBuilder:generate:generateFullAddress

func init() {
	spec := bloblang.NewPluginSpec().
		Description("Generates a randomly selected real full address that exists in the United States.").
		Param(bloblang.NewInt64Param("max_length").Default(100).Description("Specifies the maximum length for the generated data. This field ensures that the output does not exceed a certain number of characters.")).
		Param(bloblang.NewInt64Param("seed").Optional().Description("An optional seed value used to generate deterministic outputs."))

	err := bloblang.RegisterFunctionV2("generate_full_address", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
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
			res, err := generateRandomFullAddress(randomizer, maxLength)
			if err != nil {
				return nil, err
			}
			return res, nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

func NewGenerateFullAddressOptsFromConfig(config *mgmtv1alpha1.GenerateFullAddress, maxLength *int64) (*GenerateFullAddressOpts, error) {
	if config == nil {
		return NewGenerateFullAddressOpts(
			nil,
			nil,
		)
	}
	return NewGenerateFullAddressOpts(
		maxLength, nil,
	)
}

func (t *GenerateFullAddress) Generate(opts any) (any, error) {
	parsedOpts, ok := opts.(*GenerateFullAddressOpts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
	}

	return generateRandomFullAddress(parsedOpts.randomizer, parsedOpts.maxLength)
}

/* Generates a random full address from the US including street address, city, state and zipcode */
func generateRandomFullAddress(randomizer rng.Rand, maxLength int64) (string, error) {
	addresses := transformers_dataset.Addresses

	var filteredAddresses []string

	for _, address := range addresses {
		addy := fmt.Sprintf(`%s %s %s, %s`, address.Address1, address.City, address.State, address.Zipcode)
		if len(addy) <= int(maxLength) {
			filteredAddresses = append(filteredAddresses, addy)
		}
	}

	// we can't generate an address that is smaller than the max length, so attempt to generate the smallest address possible which is , if not, generate random text
	if len(filteredAddresses) == 0 {
		if maxLength < 17 {
			str, err := transformer_utils.GenerateRandomStringWithDefinedLength(randomizer, maxLength)
			if err != nil {
				return "", err
			}
			return str, nil
		} else {
			sa, err := generateRandomStreetAddress(randomizer, 5)
			if err != nil {
				return "", err
			}
			city, err := generateRandomCity(randomizer, 5)
			if err != nil {
				return "", err
			}

			state := generateRandomState(randomizer, false)

			zip := generateRandomZipcode(randomizer)

			return fmt.Sprintf(`%s %s %s, %s`, sa, city, state, zip), nil
		}
	}

	randomIndex := randomizer.Intn(len(filteredAddresses))
	return filteredAddresses[randomIndex], nil
}
