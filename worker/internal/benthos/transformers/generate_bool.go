package transformers

import (
	"math/rand"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
)

func init() {

	spec := bloblang.NewPluginSpec()

	err := bloblang.RegisterFunctionV2("generate_random_bool", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		return func() (any, error) {
			return GenerateRandomBool(), nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

// Generates a random bool value
func GenerateRandomBool() bool {

	randInt := rand.Intn(2)

	return randInt == 1

}
