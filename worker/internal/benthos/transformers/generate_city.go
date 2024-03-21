package transformers

import (
	"math/rand"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformers_dataset "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/data-sets"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

func init() {
	spec := bloblang.NewPluginSpec().Param(bloblang.NewInt64Param("max_length"))

	err := bloblang.RegisterFunctionV2("generate_city", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		maxLength, err := args.GetInt64("max_length")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {
			res, err := GenerateRandomCity(maxLength)
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

// Generates a randomly selected city that exists in the United States. Accounts for the maxLength of the column and searches for a city that is shorter than the maxLength. If not, it randomly generates a string that len(string) == maxLength
func GenerateRandomCity(maxLength int64) (string, error) {
	addresses := transformers_dataset.Addresses
	var filteredCities []string

	for _, address := range addresses {
		if len(address.City) <= int(maxLength) {
			filteredCities = append(filteredCities, address.City)
		}
	}

	if len(filteredCities) == 0 {
		city, err := transformer_utils.GenerateRandomStringWithDefinedLength(maxLength)
		if err != nil {
			return "", err
		}
		return city, nil
	}

	// -1 because addresses is an array so we don't overflow
	//nolint:all
	randomIndex := rand.Intn(len(filteredCities) - 1)

	return filteredCities[randomIndex], nil
}
