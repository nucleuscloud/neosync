package transformers

import (
	_ "embed"
	"fmt"
	"math/rand"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformers_dataset "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/data-sets"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

type Address struct {
	Address1 string `json:"address1"`
	Address2 string `json:"address2"`
	City     string `json:"city"`
	State    string `json:"state"`
	Zipcode  string `json:"zipcode"`
}

func init() {
	spec := bloblang.NewPluginSpec().Param(bloblang.NewInt64Param("max_length")).Param(bloblang.NewInt64Param("seed"))

	err := bloblang.RegisterFunctionV2("generate_street_address", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		maxLength, err := args.GetInt64("max_length")
		if err != nil {
			return nil, err
		}

		seed, err := args.GetOptionalInt64("seed")
		if err != nil {
			return nil, err
		}

		var randomizer *rand.Rand
		if seed == nil {
			randomizer = rand.New(rand.NewSource(int64(rand.Int())))
		} else {
			randomizer = rand.New(rand.NewSource(*seed))
		}

		return func() (any, error) {
			res, err := GenerateRandomStreetAddress(maxLength, randomizer)
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

/* Generates a random street address in the United States in the format <house_number> <street name> <street ending>*/
func GenerateRandomStreetAddress(maxLength int64, randomizer *rand.Rand) (string, error) {
	addresses := transformers_dataset.Addresses
	var filteredAddresses []string

	for _, address := range addresses {
		if len(address.Address1) <= int(maxLength) {
			filteredAddresses = append(filteredAddresses, address.Address1)
		}
	}

	if len(filteredAddresses) == 0 {
		if maxLength > 3 {
			hn, err := transformer_utils.GenerateRandomInt64InValueRange(1, 20, randomizer)
			if err != nil {
				return "", err
			}

			street, err := transformer_utils.GenerateRandomStringWithDefinedLength(maxLength-3, randomizer)
			if err != nil {
				return "", err
			}

			return fmt.Sprintf("%d %s", hn, street), nil
		} else {
			street, err := transformer_utils.GenerateRandomStringWithDefinedLength(maxLength, randomizer)
			if err != nil {
				return "", err
			}

			return street, nil
		}
	}

	// -1 because addresses is an array so we don't overflow
	//nolint:all
	randomIndex := rand.Intn(len(filteredAddresses) - 1)

	return filteredAddresses[randomIndex], nil
}
