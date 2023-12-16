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

	err := bloblang.RegisterFunctionV2("generate_e164_number", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		min, err := args.GetInt64("min")
		if err != nil {
			return nil, err
		}

		max, err := args.GetInt64("max")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {
			res, err := GenerateRandomE164Phone(min, max)
			return res, err
		}, nil
	})

	if err != nil {
		panic(err)
	}

}

// Generates a random phone number in e164 format and returns it as a string
func GenerateRandomE164Phone(min, max int64) (string, error) {

	if transformer_utils.GetInt64Length(min) < 9 || transformer_utils.GetInt64Length(max) > 15 {
		return "", errors.New("the length has between 9 and 15 characters long")
	}

	val, err := transformer_utils.GenerateRandomInt64WithInclusiveBounds(min, max)
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
