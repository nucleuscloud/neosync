package transformers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewInt64Param("min")).
		Param(bloblang.NewInt64Param("max"))

	err := bloblang.RegisterFunctionV2("generate_e164_phone_number", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		min, err := args.GetInt64("min")
		if err != nil {
			return nil, err
		}

		max, err := args.GetInt64("max")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {
			res, err := GenerateRandomE164PhoneNumber(min, max)
			return res, err
		}, nil
	})

	if err != nil {
		panic(err)
	}

}

/*  Generates a random phone number in e164 format in the length interval [min, max] with the min length == 9 and the max length == 15.
 */
func GenerateRandomE164PhoneNumber(min, max int64) (string, error) {

	if min < 9 || max > 15 {
		return "", errors.New("the length has between 9 and 15 characters long")
	}

	val, err := transformer_utils.GenerateRandomInt64InLengthRange(min, max)
	if err != nil {
		return "", nil
	}

	return fmt.Sprintf("+%d", val), nil
}

func ValidateE164(p string) bool {

	if len(p) >= 10 && len(p) <= 15 && strings.Contains(p, "+") {
		return true
	}
	return false
}
