package transformers

import (
	_ "embed"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformers_dataset "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/data-sets"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

var lastNames = transformers_dataset.LastNames.Names

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewStringParam("value")).Param(bloblang.NewBoolParam("preserve_length"))

	err := bloblang.RegisterFunctionV2("transform_last_name", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		value, err := args.GetString("value")
		if err != nil {
			return nil, err
		}
		preserveLength, err := args.GetBool("preserve_length")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {
			res, err := TransformLastName(value, preserveLength)
			return res, err
		}, nil

	})

	if err != nil {
		panic(err)
	}

}

// Generates a random last name which can be of either random length between [2,12] characters or as long as the input name
func TransformLastName(name string, preserveLength bool) (string, error) {

	if preserveLength {
		res, err := GenerateRandomLastNameWithLength(name)
		if err != nil {
			return "", err
		}
		return res, nil
	} else {
		res, err := GenerateRandomLastName()
		if err != nil {
			return "", err
		}
		return res, nil
	}
}

// Generates a random last name with the same length as the input last name if the length of the input last name
// is between [2,12] characters long. Otherwise, it will return a name that is 4 characters long.
func GenerateRandomLastNameWithLength(ln string) (string, error) {

	var returnValue string
	for _, v := range lastNames {
		if v.NameLength == len(ln) {
			res, err := transformer_utils.GetRandomValueFromSlice[string](v.Names)
			if err != nil {
				return "", err
			}
			returnValue = res
			break
		}
	}

	if returnValue == "" {
		res, err := transformer_utils.GetRandomValueFromSlice[string](lastNames[3].Names)
		if err != nil {
			return "", err
		}

		returnValue = res

	}

	return returnValue, nil
}
