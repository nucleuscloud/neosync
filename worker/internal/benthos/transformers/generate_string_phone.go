package transformers

import (
	"fmt"
	"strconv"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewBoolParam("include_hyphens"))

	err := bloblang.RegisterFunctionV2("generate_string_phone", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		includeHyphens, err := args.GetBool("include_hyphens")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {
			res, err := GenerateRandomPhoneNumber(includeHyphens)
			return res, err
		}, nil
	})

	if err != nil {
		panic(err)
	}

}

// generates a random phone number and returns it as a string
func GenerateRandomPhoneNumber(includeHyphens bool) (string, error) {

	defaultPhoneLength := 10

	if !includeHyphens {
		res, err := GenerateRandomPhoneNumberNoHyphens(defaultPhoneLength)
		if err != nil {
			return "", err
		}

		return res, nil

	} else {
		// only works with 10 digit-based phone numbers like in the US
		res, err := GenerateRandomPhoneNumberHyphens()
		if err != nil {
			return "", err
		}

		return res, nil

	}
}

// generates a random phone number with hyphens and returns it as a string
func GenerateRandomPhoneNumberHyphens() (string, error) {

	// only works with 10 digit-based phone numbers like in the US
	val, err := transformer_utils.GenerateRandomInt(10)

	if err != nil {
		return "", nil
	}

	areaCode := val / 10000000       // First 3 digits
	exchange := (val / 10000) % 1000 // Next 3 digits
	lineNumber := val % 10000        // Last 4 digits

	return fmt.Sprintf("%03d-%03d-%04d", areaCode, exchange, lineNumber), nil
}

// generates a random phone number of length 10 and returns it as a string
func GenerateRandomPhoneNumberNoHyphens(length int) (string, error) {

	// returns a phone number with no hyphens
	val, err := transformer_utils.GenerateRandomInt(length)
	if err != nil {
		return "", err
	}

	returnValue := strconv.FormatInt(int64(val), 10)

	return returnValue, nil
}
