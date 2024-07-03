package transformers

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformers_dataset "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/data-sets"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
)

// +neosyncTransformerBuilder:generate:generateFullAddress

func init() {
	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewInt64Param("max_length"))

	err := bloblang.RegisterFunctionV2("generate_full_address", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		maxLength, err := args.GetInt64("max_length")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {
			res, err := generateRandomFullAddress(maxLength)
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

func (t *GenerateFullAddress) Generate(opts any) (any, error) {
	parsedOpts, ok := opts.(*GenerateFullAddressOpts)
	if !ok {
		return nil, errors.New("invalid parse opts")
	}

	return generateRandomFullAddress(parsedOpts.maxLength)
}

/* Generates a random full address from the US including street address, city, state and zipcode */
func generateRandomFullAddress(maxLength int64) (string, error) {
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
			str, err := transformer_utils.GenerateRandomStringWithDefinedLength(maxLength)
			if err != nil {
				return "", err
			}
			return str, nil
		} else {
			sa, err := generateRandomStreetAddress(5)
			if err != nil {
				return "", err
			}
			city, err := generateRandomCity(5)
			if err != nil {
				return "", err
			}

			state := generateRandomState()

			zip := generateRandomZipcode()

			return fmt.Sprintf(`%s %s %s, %s`, sa, city, state, zip), nil
		}
	}

	//nolint:gosec
	randomIndex := rand.Intn(len(filteredAddresses))
	return filteredAddresses[randomIndex], nil
}
