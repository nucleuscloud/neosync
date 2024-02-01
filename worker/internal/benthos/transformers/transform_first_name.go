package transformers

import (
	_ "embed"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformers_dataset "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/data-sets"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

var firstNames = transformers_dataset.FirstNames.Names

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewInt64Param("max_length").Optional()).
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(bloblang.NewBoolParam("preserve_length"))

	err := bloblang.RegisterFunctionV2("transform_first_name", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		valuePtr, err := args.GetOptionalString("value")
		if err != nil {
			return nil, err
		}

		var value string
		if valuePtr != nil {
			value = *valuePtr
		}

		preserveLength, err := args.GetBool("preserve_length")
		if err != nil {
			return nil, err
		}

		maxLength, err := args.GetOptionalInt64("max_length")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {
			res, err := TransformFirstName(value, preserveLength, maxLength)
			return res, err
		}, nil

	})

	if err != nil {
		panic(err)
	}

}

// Generates a random first name which can be of either random length between [2,12] characters or as long as the input name
func TransformFirstName(name string, preserveLength bool, maxLength *int64) (*string, error) {

	if name == "" {
		return nil, nil
	}

	if preserveLength {
		res, err := GenerateRandomFirstNameWithLength(name)
		if err != nil {
			return nil, err
		}
		return &res, nil
	} else {
		res, err := GenerateRandomFirstName()
		if err != nil {
			return nil, err
		}

		if maxLength != nil {
			if int64(len(res)) > *maxLength {
				truncatedRes := res[:*maxLength]
				return &truncatedRes, nil
			}
		}

		return &res, nil
	}
}

// Generates a random first name with the same length as the input first name if the length of the input first name
// is between [2,12] characters long. Otherwise, it will return a name that is 4 characters long.
func GenerateRandomFirstNameWithLength(fn string) (string, error) {

	var returnValue string

	for _, v := range firstNames {
		if v.NameLength == len(fn) {
			res, err := transformer_utils.GetRandomValueFromSlice[string](v.Names)
			if err != nil {
				return "", err
			}
			returnValue = res
			break
		}
	}

	if returnValue == "" {
		res, err := transformer_utils.GetRandomValueFromSlice[string](firstNames[3].Names)
		if err != nil {
			return "", err
		}

		returnValue = res

	}

	return returnValue, nil
}
