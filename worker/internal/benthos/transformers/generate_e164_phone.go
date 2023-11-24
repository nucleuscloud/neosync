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
		Param(bloblang.NewInt64Param("length"))

	err := bloblang.RegisterFunctionV2("generate_e164_phone", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		length, err := args.GetInt64("length")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {
			res, err := GenerateRandomE164Phone(length)
			return res, err
		}, nil
	})

	if err != nil {
		panic(err)
	}

}

// Generates a random phone number in e164 format and returns it as a string
func GenerateRandomE164Phone(length int64) (string, error) {

	if length > 15 || length < 9 {
		return "", errors.New("the length has between 9 and 15 characters long")
	}

	res, err := GenerateE164FormatPhoneNumber(length)
	if err != nil {
		return "", err
	}

	return res, nil

}

// generates a random E164 phone number
func GenerateE164FormatPhoneNumber(length int64) (string, error) {

	val, err := transformer_utils.GenerateRandomInt(int(length))
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
