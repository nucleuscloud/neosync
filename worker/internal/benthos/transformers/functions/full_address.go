package neosync_benthos_transformers_functions

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
	err := bloblang.RegisterFunctionV2("fulladdresstransformer", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		return func() (any, error) {

			val, err := GenerateRandomFullAddress()

			if err != nil {
				return false, fmt.Errorf("unable to generate random state")
			}
			return val, nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

// generates a random full address from the US including street address, city, state and zipcode
func GenerateRandomFullAddress() (string, error) {

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

	fullAddress := fmt.Sprintf(`%s %s %s, %s`, randomAddress.Address1, randomAddress.City, randomAddress.State, randomAddress.Zipcode)

	return fullAddress, nil
}
