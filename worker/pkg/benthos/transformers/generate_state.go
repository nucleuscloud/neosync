package transformers

import (
	"math/rand"

	transformers_dataset "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/data-sets"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

// +neosyncTransformerBuilder:generate:generateState

func init() {
	spec := bloblang.NewPluginSpec()

	err := bloblang.RegisterFunctionV2("generate_state", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		return func() (any, error) {
			return generateRandomState(), nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

func (t *GenerateState) Generate(opts any) (any, error) {
	return generateRandomState(), nil
}

// Generates a randomly selected state that exists in the United States and returns the two-letter state code.
func generateRandomState() string {
	addresses := transformers_dataset.Addresses

	//nolint:gosec
	randomIndex := rand.Intn(len(addresses))
	return addresses[randomIndex].State
}
