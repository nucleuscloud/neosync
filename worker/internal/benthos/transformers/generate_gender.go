package transformers

import (
	"math/rand"

	"github.com/benthosdev/benthos/v4/public/bloblang"
)

func init() {
	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewBoolParam("abbreviate")).
		Param(bloblang.NewInt64Param("max_length"))

	err := bloblang.RegisterFunctionV2("generate_gender", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		ab, err := args.GetBool("abbreviate")
		if err != nil {
			return nil, err
		}

		maxLength, err := args.GetInt64("max_length")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {
			res, err := GenerateRandomGender(ab, maxLength)

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

/* Generates a randomly selected gender from a predefined list */
func GenerateRandomGender(ab bool, maxLength int64) (string, error) {
	//nolint:all
	randomInt := rand.Intn(4)

	genderMap := map[int]string{
		0: "undefined",
		1: "nonbinary",
		2: "female",
		3: "male",
	}
	// we check if the maxLength is less than 9 since our longest non-abbreviated gender is 9 digits long
	if ab || maxLength < 9 {
		return genderMap[randomInt][:1], nil
	} else {
		return genderMap[randomInt], nil
	}
}
