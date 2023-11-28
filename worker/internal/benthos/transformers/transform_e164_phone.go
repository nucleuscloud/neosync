package transformers

import (
	"fmt"
	"strings"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

var defaultE164Length = 12

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(bloblang.NewBoolParam("preserve_length"))

	err := bloblang.RegisterFunctionV2("transform_e164_phone", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

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

		return func() (any, error) {
			res, err := TransformE164Number(value, preserveLength)
			return res, err
		}, nil
	})

	if err != nil {
		panic(err)
	}

}

// Generates a random phone number and returns it as a string
func TransformE164Number(phone string, preserveLength bool) (*string, error) {

	var returnValue string

	if phone == "" {
		return nil, nil
	}

	if preserveLength {
		res, err := GenerateE164FormatPhoneNumberPreserveLength(phone)
		if err != nil {
			return nil, err
		}

		returnValue = res

	} else {

		res, err := GenerateE164FormatPhoneNumber(int64(defaultE164Length))
		if err != nil {
			return nil, err
		}
		returnValue = res
	}

	return &returnValue, nil

}

// generates a random E164 phone number and returns it as a string
func GenerateE164FormatPhoneNumberPreserveLength(number string) (string, error) {

	val := strings.Split(number, "+")

	vals, err := transformer_utils.GenerateRandomInt(len(val[1]))
	if err != nil {
		return "", nil
	}
	return fmt.Sprintf("+%d", vals), nil
}
