package neosync_transformers

import (
	"crypto/rand"
	_ "embed"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
)

func init() {

	spec := bloblang.NewPluginSpec()

	// register the function
	err := bloblang.RegisterFunctionV2("citytransformer", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		return func() (any, error) {

			val, err := GenerateRandomCity()

			if err != nil {
				return false, fmt.Errorf("unable to generate random city")
			}
			return val, nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

func GenerateRandomCity() (string, error) {

	data := struct {
		Addresses []Address `json:"addresses"`
	}{}
	if err := json.Unmarshal(addressesBytes, &data); err != nil {
		panic(err)
	}
	addresses := data.Addresses

	// -1 because addresses is an array so we don't overflow
	randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(maxIndex)-1))
	if err != nil {
		return Address{}.Address1, err
	}

	randomAddress := addresses[randomIndex.Int64()]

	return randomAddress.City, nil
}
