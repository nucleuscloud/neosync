package neosync_transformers

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
)

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewBoolParam("abbreviate"))

	// register the function

	err := bloblang.RegisterFunctionV2("gendertransformer", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		ab, err := args.GetBool("abbreviate")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {

			val, err := GenerateRandomGender(ab)

			if err != nil {
				return false, fmt.Errorf("unable to generate random gender")
			}
			return val, nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

// generates a randomly selected gender
func GenerateRandomGender(ab bool) (string, error) {

	randomInt, err := rand.Int(rand.Reader, big.NewInt(4))
	if err != nil {
		return "", err
	}

	var gender string

	switch randomInt.Int64() {
	case 0:
		gender = "undefined"
	case 1:
		gender = "nonbinary"
	case 2:
		gender = "female"
	case 3:
		gender = "male"
	}

	if ab {
		gender = gender[:1]
	}

	return gender, nil
}
