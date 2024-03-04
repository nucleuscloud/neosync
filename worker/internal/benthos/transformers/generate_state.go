package transformers

import (
	"math/rand"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformers_dataset "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/data-sets"
)

func init() {
	spec := bloblang.NewPluginSpec()

	err := bloblang.RegisterFunctionV2("generate_state", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		return func() (any, error) {
			return GenerateRandomState(), nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

// Generates a randomly selected state that exists in the United States and returns the two-letter state code.
func GenerateRandomState() string {
	addresses := transformers_dataset.Addresses

	// -1 because addresses is an array so we don't overflow
	//nolint:all
	randomIndex := rand.Intn(len(addresses) - 1)

	return addresses[randomIndex].State
}
