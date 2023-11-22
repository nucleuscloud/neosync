package transformers

import (
	"math/rand"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformers_dataset "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/data-sets"
)

func init() {

	spec := bloblang.NewPluginSpec()

	err := bloblang.RegisterFunctionV2("generate_random_city", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		return func() (any, error) {
			return GenerateRandomCity(), nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

// Generates a randomly selected city that exists in the United States.
func GenerateRandomCity() string {

	addresses := transformers_dataset.Addresses

	// -1 because addresses is an array so we don't overflow
	randomIndex := rand.Intn(len(addresses) - 1)

	return addresses[randomIndex].City
}
