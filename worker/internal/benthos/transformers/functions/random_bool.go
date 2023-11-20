package neosync_transformers

import (
	"crypto/rand"
	"fmt"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
)

func init() {

	spec := bloblang.NewPluginSpec()

	// register the function

	err := bloblang.RegisterFunctionV2("randombooltransformer", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		return func() (any, error) {

			val, err := GenerateRandomBool()

			if err != nil {
				return false, fmt.Errorf("unable to generate random bool")
			}
			return val, nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

func GenerateRandomBool() (bool, error) {

	// Create a random source using crypto/rand
	source := rand.Reader

	// read a random byte from the source
	buf := make([]byte, 1)

	_, err := source.Read(buf)
	if err != nil {
		return false, fmt.Errorf("unable to generate a random boolean")
	}

	// read least sig bit from byte and return bool
	return buf[0]&1 == 1, nil

}
