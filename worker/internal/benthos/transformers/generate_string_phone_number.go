package transformers

import (
	"fmt"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

func init() {
	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewInt64Param("min")).
		Param(bloblang.NewInt64Param("max"))

	err := bloblang.RegisterFunctionV2("generate_string_phone_number", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		min, err := args.GetInt64("min")
		if err != nil {
			return nil, err
		}

		max, err := args.GetInt64("max")
		if err != nil {
			return nil, err
		}
		return func() (any, error) {
			res, err := GenerateStringPhoneNumber(min, max)
			if err != nil {
				return nil, fmt.Errorf("unable to run generate_string_phone_number: %w", err)
			}
			return res, nil
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

/*  Generates a string phone number in the length interval [min, max] with the min length == 9 and the max length == 15.
 */
func GenerateStringPhoneNumber(min, max int64) (string, error) {
	min = transformer_utils.MaxInt(9, min)
	max = transformer_utils.MinInt(15, max)

	val, err := transformer_utils.GenerateRandomInt64InLengthRange(min, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d", val), nil
}
