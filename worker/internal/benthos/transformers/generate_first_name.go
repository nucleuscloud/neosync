package transformers

import (
	_ "embed"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
)

func init() {

	spec := bloblang.NewPluginSpec()

	err := bloblang.RegisterFunctionV2("generate_first_name", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		return func() (any, error) {
			return GenerateRandomFirstName(int64(23))
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

/* Generates a random first name with a randomly selected length between [2,12] characters */
func GenerateRandomFirstName(maxLength int64) (string, error) {

	// var returnValue string

	// var nameLengths []int

	// var firstNames = transformers_dataset.FirstNames.Names

	// for _, v := range firstNames {
	// 	nameLengths = append(nameLengths, v.NameLength)
	// }

	// randomNameLengthVal, err := transformer_utils.GetRandomValueFromSlice[int](nameLengths)
	// if err != nil {
	// 	return "", err
	// }

	// for _, v := range firstNames {
	// 	if v.NameLength == randomNameLengthVal {
	// 		res, err := transformer_utils.GetRandomValueFromSlice[string](v.Names)
	// 		if err != nil {
	// 			return "", err
	// 		}
	// 		returnValue = res
	// 	}
	// }

	// return returnValue, nil

	return "hello", nil
}
