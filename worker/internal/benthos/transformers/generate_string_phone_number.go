package transformers

import (
	"errors"
	"fmt"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

func init() {
	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewInt64Param("min")).
		Param(bloblang.NewInt64Param("max")).
		Param(bloblang.NewInt64Param("max_length"))

	err := bloblang.RegisterFunctionV2("generate_string_phone_number", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		min, err := args.GetInt64("min")
		if err != nil {
			return nil, err
		}

		max, err := args.GetInt64("max")
		if err != nil {
			return nil, err
		}

		maxLength, err := args.GetInt64("max_length")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {
			res, err := GenerateStringPhoneNumber(min, max, maxLength)
			return res, err
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

/*  Generates a string phone number in the length interval [min, max] with the min length == 9 and the max length == 15.
 */
func GenerateStringPhoneNumber(min, max, maxLength int64) (string, error) {
	if min < 9 || max > 15 {
		return "", errors.New("the length has between 9 and 15 characters long")
	}

	if max > maxLength {
		val, err := transformer_utils.GenerateRandomInt64InLengthRange(min, maxLength-1)
		if err != nil {
			return "", nil
		}

		return fmt.Sprintf("%d", val), nil
	} else {
		val, err := transformer_utils.GenerateRandomInt64InLengthRange(min, max)
		if err != nil {
			return "", nil
		}

		return fmt.Sprintf("%d", val), nil
	}
}
