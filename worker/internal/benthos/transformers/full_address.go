package transformers

import (
	_ "embed"
	"fmt"
	"math/rand"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformers_dataset "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/data-sets"
)

func init() {

	spec := bloblang.NewPluginSpec()

	// register the function
	err := bloblang.RegisterFunctionV2("generate_random_full_address", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		return func() (any, error) {
			return GenerateRandomFullAddress(), nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

// generates a random full address from the US including street address, city, state and zipcode
func GenerateRandomFullAddress() string {

	addresses := transformers_dataset.Addresses

	// -1 because addresses is an array so we don't overflow
	randomIndex := rand.Intn(len(addresses) - 1)

	randomAddress := addresses[randomIndex]

	fullAddress := fmt.Sprintf(`%s %s %s, %s`, randomAddress.Address1, randomAddress.City, randomAddress.State, randomAddress.Zipcode)

	return fullAddress
}
