package transformers

import (
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	transformers_dataset "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/data-sets"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
)

// +neosyncTransformerBuilder:generate:generateFullAddress

func init() {
	spec := bloblang.NewPluginSpec().
		Description("Generates a randomly selected real full address that exists in the United States.").
		Category("string").
		Param(bloblang.NewInt64Param("max_length").Default(100).Description("Specifies the maximum length for the generated data. This field ensures that the output does not exceed a certain number of characters.")).
		Param(bloblang.NewInt64Param("seed").Optional().Description("An optional seed value used to generate deterministic outputs."))

	err := bloblang.RegisterFunctionV2(
		"generate_full_address",
		spec,
		func(args *bloblang.ParsedParams) (bloblang.Function, error) {
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
		},
	)
	if err != nil {
		panic(err)
	}
}

func NewGenerateFullAddressOptsFromConfig(
	config *mgmtv1alpha1.GenerateFullAddress,
	maxLength *int64,
) (*GenerateFullAddressOpts, error) {
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
	zipcode, err := generateRandomZipcode(randomizer)
	if err != nil {
		return "", err
	}
	state, err := generateRandomState(randomizer, false)
	if err != nil {
		return "", err
	}

	// todo: we could further generate a cache for this to skip having to potentially re-calculate this every time for the given length
	// we have a finite set of zipcodes and states so we basically know the max length for the city and street address for each generated permutation.
	remainder := int64(int(maxLength) - len(state) - len(zipcode) - 4) // -4 for spaces and comma
	if remainder <= 0 {
		return "", fmt.Errorf(
			"the state and zipcode combined are longer than the max length allowed",
		)
	}

	maxCityIdx, maxAddr1Idx := transformer_utils.FindClosestPair(
		transformers_dataset.Address_CityIndices,
		transformers_dataset.Address_Address1Indices,
		remainder,
	)
	if maxCityIdx == -1 || maxAddr1Idx == -1 {
		randStr, err := transformer_utils.GenerateRandomStringWithInclusiveBounds(
			randomizer,
			1,
			remainder,
		)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf(`%s %s, %s`, randStr, state, zipcode), nil
	}

	city, err := generateRandomCity(
		randomizer,
		transformers_dataset.Address_CityIndices[maxCityIdx],
	)
	if err != nil {
		return "", err
	}

	street, err := generateRandomStreetAddress(
		randomizer,
		transformers_dataset.Address_Address1Indices[maxAddr1Idx],
	)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`%s %s %s, %s`, street, city, state, zipcode), nil
}
