package transformers

import (
	"math/rand"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformers_dataset "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/data-sets"
)

// +neosyncTransformerBuilder:generate:generateZipcode

func init() {
	spec := bloblang.NewPluginSpec()

	err := bloblang.RegisterFunctionV2("generate_zipcode", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		return func() (any, error) {
			return generateRandomZipcode(), nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

func (t *GenerateZipcode) Generate(opts any) (any, error) {
	return generateRandomZipcode(), nil
}

// Generates a randomly selected zip code that exists in the United States.
func generateRandomZipcode() string {
	addresses := transformers_dataset.Addresses
	//nolint:gosec
	randomIndex := rand.Intn(len(addresses))
	return addresses[randomIndex].Zipcode
}
