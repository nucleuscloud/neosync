package neosync_benthos_transformers_functions

import (
	"fmt"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

func init() {

	spec := bloblang.NewPluginSpec()

	// register the function
	err := bloblang.RegisterFunctionV2("ssntransformer", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		return func() (any, error) {

			val, err := GenerateRandomSSN()

			if err != nil {
				return false, fmt.Errorf("unable to generate random ssn")
			}
			return val, nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

func GenerateRandomSSN() (string, error) {

	val, err := transformer_utils.GenerateRandomInt(9)
	if err != nil {
		return "", err
	}

	returnVal := fmt.Sprintf("%03d-%02d-%04d", val/1000000, (val/10000)%100, val%10000)

	return returnVal, nil
}
