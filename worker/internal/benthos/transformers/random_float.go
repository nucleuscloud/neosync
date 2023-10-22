package neosync_transformers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
)

const defaultLenBeforeDecimals = 2
const defaultLenAfterDecimals = 3

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewBoolParam("preserve_length")).
		Param(bloblang.NewInt64Param("digits_before_decimal")).
		Param(bloblang.NewInt64Param("digits_after_decimal"))

	// register the plugin
	err := bloblang.RegisterMethodV2("randomfloattransformer", spec, func(args *bloblang.ParsedParams) (bloblang.Method, error) {

		preserveLength, err := args.GetBool("preserve_length")
		if err != nil {
			return nil, err
		}

		digitsBeforeDecimal, err := args.GetInt64("digits_before_decimal")
		if err != nil {
			return nil, err
		}

		digitsAfterDecimal, err := args.GetInt64("digits_after_decimal")
		if err != nil {
			return nil, err
		}

		return bloblang.Float64Method(func(i float64) (any, error) {
			res, err := ProcessRandomFloat(i, preserveLength, digitsBeforeDecimal, digitsAfterDecimal)
			return res, err
		}), nil
	})

	if err != nil {
		panic(err)
	}

}

// main transformer logic goes here
func ProcessRandomFloat(i float64, preserveLength bool, digitsBeforeDecimal, digitsAfterDecimal int64) (float64, error) {

	var returnValue float64

	fLen := GetFloatLength(i)

	if digitsBeforeDecimal < 0 || digitsAfterDecimal < 0 {
		return 0.0, fmt.Errorf("digitsBefore and digitsAfter must be non-negative")
	}

	if preserveLength {

		bd, err := GenerateRandomInt(int64(fLen.DigitsBeforeDecimalLength))
		if err != nil {
			return 0, fmt.Errorf("unable to generate a random before digits integer")
		}

		ad, err := GenerateRandomInt(int64(fLen.DigitsAfterDecimalLength))

		if err != nil {
			return 0, fmt.Errorf("unable to generate a random after digits integer")
		}

		combinedStr := fmt.Sprintf("%d.%d", bd, ad)

		result, err := strconv.ParseFloat(combinedStr, 64)
		if err != nil {
			return 0, fmt.Errorf("unable to convert string to float")
		}

		returnValue = result
	} else {

		bd, err := GenerateRandomInt(int64(defaultLenBeforeDecimals))
		if err != nil {
			return 0, fmt.Errorf("unable to generate a random before digits integer")
		}

		ad, err := GenerateRandomInt(int64(defaultLenAfterDecimals))

		// generate a new number if it ends in a zero so that the trailing zero doesn't get stripped and return
		// a value that is shorter than what the user asks for. This happens in the when we convert the string to a float64
		for {
			if !LastDigitAZero(ad) {
				break // Exit the loop when i is greater than or equal to 5
			}
			ad, err = GenerateRandomInt(int64(fLen.DigitsAfterDecimalLength))

			if err != nil {
				return 0, fmt.Errorf("unable to generate a random int64 to convert to a float")
			}
		}

		if err != nil {
			return 0, fmt.Errorf("unable to generate a random after digits integer")
		}

		if err != nil {
			return 0, fmt.Errorf("unable to generate a random string with length")
		}

		combinedStr := fmt.Sprintf("%d.%d", bd, ad)

		result, err := strconv.ParseFloat(combinedStr, 64)
		if err != nil {
			return 0, fmt.Errorf("unable to convert string to float")
		}

		returnValue = result

	}

	return returnValue, nil
}

type FloatLength struct {
	DigitsBeforeDecimalLength int
	DigitsAfterDecimalLength  int
}

func GetFloatLength(i float64) *FloatLength {
	// Convert the int64 to a string
	str := fmt.Sprintf("%g", i)

	parsed := strings.Split(str, ".")

	return &FloatLength{
		DigitsBeforeDecimalLength: len(parsed[0]),
		DigitsAfterDecimalLength:  len(parsed[1]),
	}
}

func LastDigitAZero(n int64) bool {
	// Convert the int64 to a string
	str := strconv.FormatInt(n, 10)

	// Check if the string is empty or if the last character is '0'
	if len(str) > 0 && str[len(str)-1] == '0' {
		return true
	}

	return false
}
