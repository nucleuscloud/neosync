package transformers

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewBoolParam("e164_format")).
		Param(bloblang.NewBoolParam("include_hyphens"))

	// register the plugin
	err := bloblang.RegisterFunctionV2("generate_string_phone", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		includeHyphens, err := args.GetBool("include_hyphens")
		if err != nil {
			return nil, err
		}

		e164, err := args.GetBool("e164_format")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {
			res, err := GenerateRandomPhoneNumber(e164, includeHyphens)
			return res, err
		}, nil
	})

	if err != nil {
		panic(err)
	}

}

// generates a random phone number and returns it as a string
func GenerateRandomPhoneNumber(e164Format, includeHyphens bool) (string, error) {

	if e164Format && includeHyphens {
		return "", errors.New("E164 phone numbers cannot have hyphens")
	}

	if !includeHyphens && !e164Format {
		res, err := GenerateRandomPhoneNumberNoHyphens()
		if err != nil {
			return "", err
		}

		return res, nil

	} else if includeHyphens && !e164Format {
		// only works with 10 digit-based phone numbers like in the US
		res, err := GenerateRandomPhoneNumberHyphens()
		if err != nil {
			return "", err
		}

		return res, nil

	} else {

		// outputs in e164 format -> for ex. +873104859612, regex: ^\+[1-9]\d{1,14}$
		res, err := GenerateRandomE164FormatPhoneNumber()
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

// generates a random E164 phone number between 10 and 15 digits long and returns it as a string
func GenerateRandomE164FormatPhoneNumber() (string, error) {

	val, err := transformer_utils.GenerateRandomInt(10)
	if err != nil {
		return "", nil
	}
	return fmt.Sprintf("+%d", val), nil

}

// generatea a random phone number of length 10 and returns it as a string
func GenerateRandomPhoneNumberNoHyphens() (string, error) {

	// returns a phone number with no hyphens
	val, err := transformer_utils.GenerateRandomInt(10)
	if err != nil {
		return "", err
	}

	returnValue := strconv.FormatInt(int64(val), 10)

	return returnValue, nil
}
