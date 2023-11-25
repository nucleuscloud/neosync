package transformers

import (
	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

func init() {

	spec := bloblang.NewPluginSpec()

	err := bloblang.RegisterFunctionV2("generate_int64_phone", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		return func() (any, error) {
			res, err := GenerateRandomIntPhoneNumber()
			return res, err
		}, nil
	})

	if err != nil {
		panic(err)
	}

}

// Generates a random 10 digit phone number and returns it as an int64
func GenerateRandomIntPhoneNumber() (int64, error) {

	res, err := transformer_utils.GenerateRandomInt(10)

	if err != nil {
		return 0, err
	}

	return int64(res), nil
}
