package transformers

import (
	_ "embed"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

func init() {

	spec := bloblang.NewPluginSpec().Param(bloblang.NewInt64Param("max_length"))

	err := bloblang.RegisterFunctionV2("generate_first_name", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		maxLength, err := args.GetInt64("max_length")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {
			return GenerateRandomFirstName(maxLength)
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

/* Generates a random first name with a randomly selected length between [2,12] characters */
func GenerateRandomFirstName(maxLength int64) (string, error) {

	if maxLength < 12 && maxLength >= 2 {
		names := firstNames[maxLength]
		res, err := transformer_utils.GetRandomValueFromSlice[string](names)
		if err != nil {
			return "", err
		}
		return res, nil
	} else {
		randInd, err := transformer_utils.GenerateRandomInt64InValueRange(2, 12)
		if err != nil {
			return "", err
		}

		names := firstNames[randInd]
		res, err := transformer_utils.GetRandomValueFromSlice[string](names)
		if err != nil {
			return "", err
		}
		return res, nil

	}
}
