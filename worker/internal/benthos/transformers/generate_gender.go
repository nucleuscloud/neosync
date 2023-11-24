package transformers

import (
	"math/rand"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
)

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewBoolParam("abbreviate"))

	err := bloblang.RegisterFunctionV2("generate_random_gender", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		ab, err := args.GetBool("abbreviate")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {

			res, err := GenerateRandomGender(ab)

			if err != nil {
				return false, err
			}
			return res, nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

// generates a randomly selected gender
func GenerateRandomGender(ab bool) (string, error) {

	randomInt := rand.Intn(4)

	var gender string

	switch randomInt {
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
