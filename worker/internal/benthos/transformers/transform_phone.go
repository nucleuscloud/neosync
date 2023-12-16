package transformers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(bloblang.NewBoolParam("preserve_length")).
		Param(bloblang.NewBoolParam("include_hyphens"))

	err := bloblang.RegisterFunctionV2("transform_phone", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

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

		includeHyphens, err := args.GetBool("include_hyphens")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {
			res, err := TransformPhoneNumber(value, preserveLength, includeHyphens)
			return res, err
		}, nil
	})

	if err != nil {
		panic(err)
	}

}

// Generates a random phone number and returns it as a string
func TransformPhoneNumber(value string, preserveLength, includeHyphens bool) (*string, error) {

	var returnValue string

	if value == "" {
		return nil, nil
	}

	if preserveLength && includeHyphens && len(value) != 10 {
		return nil, fmt.Errorf("can only preserve the length of the input phone number and include hyphens if the length of the phone number is 10")
	} else {
		// only works with 10 digit-based phone numbers like in the US
		res, err := GenerateRandomPhoneNumberWithHyphens()
		if err != nil {
			return nil, err
		}
		returnValue = res
	}

	if preserveLength && !includeHyphens {
		res, err := GeneratePhoneNumberPreserveLengthNoHyphensNotE164(value)
		if err != nil {
			return nil, err
		}

		returnValue = res

	} else if !preserveLength && includeHyphens {
		// only works with 10 digit-based phone numbers like in the US
		res, err := GenerateRandomPhoneNumberWithHyphens()
		if err != nil {
			return nil, err
		}

		returnValue = res

	} else if !preserveLength && !includeHyphens {

		res, err := GenerateRandomPhoneNumberWithNoHyphens()
		if err != nil {
			return nil, err
		}

		returnValue = res
	}

	return &returnValue, nil

}

// generates a random phone number by presrving the length of the input phone number, removes hyphens and is not in e164 format
func GeneratePhoneNumberPreserveLengthNoHyphensNotE164(number string) (string, error) {

	if strings.Contains(number, "-") { // checks if input phone number has hyphens
		number = strings.ReplaceAll(number, "-", "")
	}
	length := int64(len(number))

	val, err := transformer_utils.GenerateRandomInt64WithInclusiveBounds(length, length)

	if err != nil {
		return "", nil
	}

	returnValue := strconv.FormatInt(val, 10)

	return returnValue, nil
}

// generates a random phone number with hyphens and returns it as a string
func GenerateRandomPhoneNumberWithHyphens() (string, error) {

	// only works with 10 digit-based phone numbers like in the US

	val, err := transformer_utils.GenerateRandomInt64WithInclusiveBounds(defaultPhoneNumberLength, defaultPhoneNumberLength)

	if err != nil {
		return "", nil
	}

	areaCode := val / 10000000       // First 3 digits
	exchange := (val / 10000) % 1000 // Next 3 digits
	lineNumber := val % 10000        // Last 4 digits

	return fmt.Sprintf("%03d-%03d-%04d", areaCode, exchange, lineNumber), nil
}

// generatea a random phone number of length 10 and returns it as a string
func GenerateRandomPhoneNumberWithNoHyphens() (string, error) {

	// returns a phone number with no hyphens
	val, err := transformer_utils.GenerateRandomInt64WithInclusiveBounds(defaultPhoneNumberLength, defaultPhoneNumberLength)
	if err != nil {
		return "", err
	}

	returnValue := strconv.FormatInt(val, 10)

	return returnValue, nil
}
