package transformers

import (
	"errors"
	"fmt"
	"math/rand"

	transformers_dataset "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/data-sets"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
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
		Param(bloblang.NewInt64Param("max_length").Description("Specifies the maximum length for the generated data. This field ensures that the output does not exceed a certain number of characters."))

	err := bloblang.RegisterFunctionV2("generate_street_address", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		maxLength, err := args.GetInt64("max_length")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {
			res, err := generateRandomStreetAddress(maxLength)
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

func (t *GenerateStreetAddress) Generate(opts any) (any, error) {
	parsedOpts, ok := opts.(*GenerateStreetAddressOpts)
	if !ok {
		return nil, errors.New("invalid parse opts")
	}

	return generateRandomStreetAddress(parsedOpts.maxLength)
}

/* Generates a random street address in the United States in the format <house_number> <street name> <street ending>*/
func generateRandomStreetAddress(maxLength int64) (string, error) {
	addresses := transformers_dataset.Addresses
	var filteredAddresses []string

	for _, address := range addresses {
		if len(address.Address1) <= int(maxLength) {
			filteredAddresses = append(filteredAddresses, address.Address1)
		}
	}

	if len(filteredAddresses) == 0 {
		if maxLength > 3 {
			hn, err := transformer_utils.GenerateRandomInt64InValueRange(1, 20)
			if err != nil {
				return "", err
			}

			street, err := transformer_utils.GenerateRandomStringWithDefinedLength(maxLength - 3)
			if err != nil {
				return "", err
			}

			return fmt.Sprintf("%d %s", hn, street), nil
		} else {
			street, err := transformer_utils.GenerateRandomStringWithDefinedLength(maxLength)
			if err != nil {
				return "", err
			}

			return street, nil
		}
	}

	//nolint:gosec
	randomIndex := rand.Intn(len(filteredAddresses))
	return filteredAddresses[randomIndex], nil
}
