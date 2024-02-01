package transformers

import (
	_ "embed"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
)

func init() {

	spec := bloblang.NewPluginSpec()

	err := bloblang.RegisterFunctionV2("generate_last_name", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		return func() (any, error) {
			res, err := GenerateRandomLastName()
			return res, err
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

/* Generates a random last name with a randomly selected length between [2,12] characters */
func GenerateRandomLastName() (string, error) {

	// var returnValue string

	// var nameLengths []int

	// var lastNames = transformers_dataset.LastNames

	// for _, v := range lastNames {
	// 	nameLengths = append(nameLengths, v.NameLength)
	// }

	// randomNameLengthVal, err := transformer_utils.GetRandomValueFromSlice[int](nameLengths)
	// if err != nil {
	// 	return "", err
	// }

	// for _, v := range lastNames {
	// 	if v.NameLength == randomNameLengthVal {
	// 		res, err := transformer_utils.GetRandomValueFromSlice[string](v.Names)
	// 		if err != nil {
	// 			return "", err
	// 		}
	// 		returnValue = res
	// 	}
	// }

	// return returnValue, nil

	return "ewe", nil
}
