package transformers

import (
	"fmt"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewInt64Param("min")).
		Param(bloblang.NewInt64Param("max"))

	err := bloblang.RegisterFunctionV2("generate_string", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		min, err := args.GetInt64("min")
		if err != nil {
			return nil, err
		}

		max, err := args.GetInt64("max")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {
			res, err := GenerateRandomString(min, max)
			return res, err
		}, nil
	})

	if err != nil {
		panic(err)
	}

}

// main transformer logic goes here
func GenerateRandomString(min, max int64) (string, error) {

	var returnValue string

	val, err := transformer_utils.GenerateRandomStringWithInclusiveBounds(min, max)

	if err != nil {
		return "", fmt.Errorf("unable to generate a random string with length")
	}

	returnValue = val

	return returnValue, nil
}
