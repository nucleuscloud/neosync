package neosync_transformers

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
		Param(bloblang.NewStringParam("phone").Optional()).
		Param(bloblang.NewBoolParam("preserve_length").Optional()).
		Param(bloblang.NewBoolParam("e164_format").Optional()).
		Param(bloblang.NewBoolParam("include_hyphens").Optional())

	// register the plugin
	err := bloblang.RegisterFunctionV2("phonetransformer", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		phonePtr, err := args.GetOptionalString("phone")
		if err != nil {
			return nil, err
		}
		var phone string
		if phonePtr != nil {
			phone = *phonePtr
		}

		preserveLengthPtr, err := args.GetOptionalBool("preserve_length")
		if err != nil {
			return nil, err
		}

		var preserveLength bool
		if preserveLengthPtr != nil {
			preserveLength = *preserveLengthPtr
		}

		includeHyphensPtr, err := args.GetOptionalBool("include_hyphens")
		if err != nil {
			return nil, err
		}

		var includeHyphens bool
		if includeHyphensPtr != nil {
			includeHyphens = *includeHyphensPtr
		}

		e164Ptr, err := args.GetOptionalBool("e164_format")
		if err != nil {
			return nil, err
		}

		var e164 bool
		if e164Ptr != nil {
			e164 = *e164Ptr
		}

		return func() (any, error) {
			res, err := GeneratePhoneNumber(phone, preserveLength, e164, includeHyphens)
			return res, err
		}, nil
	})

	if err != nil {
		panic(err)
	}

}

// generates a random phone number and returns it as a string
func GeneratePhoneNumber(phone string, preserveLength, e164Format, includeHyphens bool) (string, error) {

	if preserveLength && includeHyphens {
		return "", fmt.Errorf("the preserve length param cannot be true if the include hyphens is true")
	}

	if phone != "" {

		if preserveLength && !includeHyphens && !e164Format {
			res, err := GeneratePhoneNumberPreserveLengthNoHyphensNotE164(phone)
			if err != nil {
				return "", err
			}

			return res, nil

		} else if !preserveLength && includeHyphens && !e164Format {
			// only works with 10 digit-based phone numbers like in the US
			res, err := GenerateRandomPhoneNumberWithHyphens()
			if err != nil {
				return "", err
			}

			return res, nil

		} else if !preserveLength && !includeHyphens && e164Format {

			/* outputs in e164 format -> for ex. +873104859612, regex: ^\+[1-9]\d{1,14}$ */
			res, err := GenerateE164FormatPhoneNumber()
			if err != nil {
				return "", err
			}

			return res, nil
		} else if e164Format && preserveLength && !includeHyphens {

			res, err := GenerateE164FormatPhoneNumberPreserveLength(phone)
			if err != nil {
				return "", err
			}

			return res, nil

		} else {

			res, err := GenerateRandomPhoneNumberWithNoHyphens()
			if err != nil {
				return "", err
			}

			return res, nil

		}

	} else {

		res, err := GenerateRandomPhoneNumberWithNoHyphens()
		if err != nil {
			return "", err
		}

		return res, nil

	}

}

// generates a random phone number by presrving the length of the input phone number, removes hyphens and is not in e164 format
func GeneratePhoneNumberPreserveLengthNoHyphensNotE164(number string) (string, error) {

	if strings.Contains(number, "-") { // checks if input phone number has hyphens
		number = strings.ReplaceAll(number, "-", "")
	}

	val, err := transformer_utils.GenerateRandomInt(int64(len(number)))

	if err != nil {
		return "", nil
	}

	returnValue := strconv.FormatInt(val, 10)

	return returnValue, nil
}

// generates a random phone number with hyphens and returns it as a string
func GenerateRandomPhoneNumberWithHyphens() (string, error) {

	// only works with 10 digit-based phone numbers like in the US

	val, err := transformer_utils.GenerateRandomInt(int64(10))

	if err != nil {
		return "", nil
	}

	areaCode := val / 10000000       // First 3 digits
	exchange := (val / 10000) % 1000 // Next 3 digits
	lineNumber := val % 10000        // Last 4 digits

	return fmt.Sprintf("%03d-%03d-%04d", areaCode, exchange, lineNumber), nil
}

// generates a random E164 phone number between 10 and 15 digits long and returns it as a string
func GenerateE164FormatPhoneNumber() (string, error) {

	val, err := transformer_utils.GenerateRandomInt(int64(10))
	if err != nil {
		return "", nil
	}
	return fmt.Sprintf("+%d", val), nil

}

// generates a random E164 phone number and returns it as a string
func GenerateE164FormatPhoneNumberPreserveLength(number string) (string, error) {

	val := strings.Split(number, "+")

	vals, err := transformer_utils.GenerateRandomInt(int64(len(val[1])))
	if err != nil {
		return "", nil
	}
	return fmt.Sprintf("+%d", vals), nil
}

// generatea a random phone number of length 10 and returns it as a string
func GenerateRandomPhoneNumberWithNoHyphens() (string, error) {

	// returns a phone number with no hyphens
	val, err := transformer_utils.GenerateRandomInt(int64(10))
	if err != nil {
		return "", err
	}

	returnValue := strconv.FormatInt(val, 10)

	return returnValue, nil
}

func ValidateE164(p string) bool {

	if len(p) >= 10 && len(p) <= 15 && strings.Contains(p, "+") {
		return true
	}
	return false
}
