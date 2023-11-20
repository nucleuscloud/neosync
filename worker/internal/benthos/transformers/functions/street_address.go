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

// the number of addresses in the file
const maxIndex = 1000

var (
	//go:embed data-sets/addresses.json
	addressesBytes []byte
)

type Address struct {
	Address1 string `json:"address1"`
	Address2 string `json:"address2"`
	City     string `json:"city"`
	State    string `json:"state"`
	Zipcode  string `json:"zipcode"`
}

func init() {

	spec := bloblang.NewPluginSpec()

	// register the function
	err := bloblang.RegisterFunctionV2("streetaddresstransformer", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		return func() (any, error) {

			val, err := GenerateRandomStreetAddress()

			if err != nil {
				return false, fmt.Errorf("unable to generate random address")
			}
			return val, nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

func GenerateRandomStreetAddress() (string, error) {

	data := struct {
		Addresses []Address `json:"addresses"`
	}{}
	if err := json.Unmarshal(addressesBytes, &data); err != nil {
		return "", err
	}
	addresses := data.Addresses

	// -1 because addresses is an array so we don't overflow
	randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(maxIndex)-1))
	if err != nil {
		return Address{}.Address1, err
	}

	randomAddress := addresses[randomIndex.Int64()]

	return randomAddress.Address1, nil
}
