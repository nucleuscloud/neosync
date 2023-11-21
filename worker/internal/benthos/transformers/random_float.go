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
		Param(bloblang.NewFloat64Param("value").Optional()).
		Param(bloblang.NewBoolParam("preserve_length").Optional()).
		Param(bloblang.NewInt64Param("digits_before_decimal").Optional()).
		Param(bloblang.NewInt64Param("digits_after_decimal").Optional())

	// register the plugin
	err := bloblang.RegisterFunctionV2("randomfloattransformer", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		valuePtr, err := args.GetOptionalFloat64("value")
		if err != nil {
			return nil, err
		}

		var value float64
		if valuePtr != nil {
			value = *valuePtr
		}

		preserveLengthPtr, err := args.GetOptionalBool("preserve_length")
		if err != nil {
			return nil, err
		}

		var preserveLength bool
		if preserveLengthPtr != nil {
			preserveLength = *preserveLengthPtr
		}

		digitsBeforeDecimalPtr, err := args.GetOptionalInt64("digits_before_decimal")
		if err != nil {
			return nil, err
		}

		var digitsBeforeDecimal int64
		if digitsBeforeDecimalPtr != nil {
			digitsBeforeDecimal = *digitsBeforeDecimalPtr
		}

		digitsAfterDecimalPtr, err := args.GetOptionalInt64("digits_after_decimal")
		if err != nil {
			return nil, err
		}

		var digitsAfterDecimal int64
		if digitsAfterDecimalPtr != nil {
			digitsAfterDecimal = *digitsAfterDecimalPtr
		}

		return func() (any, error) {
			res, err := GenerateRandomFloat(value, preserveLength, digitsAfterDecimal, digitsBeforeDecimal)
			return res, err

		}, nil
	})

	if err != nil {
		panic(err)
	}

}

func GenerateRandomFloat(value float64, preserveLength bool, digitsAfterDecimal, digitsBeforeDecimal int64) (float64, error) {

	if value != 0 {
		if preserveLength {
			fLen := GetFloatLength(value)
			res, err := GenerateRandomFloatWithDefinedLength(int64(fLen.DigitsBeforeDecimalLength), int64(fLen.DigitsAfterDecimalLength))
			return res, err
		} else {
			res, err := GenerateRandomFloatWithDefinedLength(digitsBeforeDecimal, digitsAfterDecimal)
			return res, err
		}
	} else {
		res, err := GenerateRandomFloatWithRandomLength()
		return res, err
	}
}

func GenerateRandomFloatWithRandomLength() (float64, error) {

	var returnValue float64

	bd, err := transformer_utils.GenerateRandomInt(int64(3))
	if err != nil {
		return 0, fmt.Errorf("unable to generate a random before digits integer")
	}

	ad, err := transformer_utils.GenerateRandomInt(int64(3))

	for {
		if !transformer_utils.IsLastDigitZero(ad) {
			break
		}
		ad, err = transformer_utils.GenerateRandomInt(int64(3))

		if err != nil {
			return 0, fmt.Errorf("unable to generate a random int64 to convert to a float")
		}
	}

	if err != nil {
		return 0, fmt.Errorf("unable to generate a random after digits integer")
	}

	combinedStr := fmt.Sprintf("%d.%d", bd, ad)

	result, err := strconv.ParseFloat(combinedStr, 64)
	if err != nil {
		return 0, fmt.Errorf("unable to convert string to float")
	}

	returnValue = result

	return returnValue, nil
}

func GenerateRandomFloatWithDefinedLength(digitsBeforeDecimal, digitsAfterDecimal int64) (float64, error) {

	var returnValue float64

	bd, err := transformer_utils.GenerateRandomInt(digitsBeforeDecimal)
	if err != nil {
		return 0, fmt.Errorf("unable to generate a random before digits integer")
	}

	ad, err := transformer_utils.GenerateRandomInt(digitsAfterDecimal)

	// generate a new number if it ends in a zero so that the trailing zero doesn't get stripped and return
	// a value that is shorter than what the user asks for. This happens in when we convert the string to a float64
	for {
		if !transformer_utils.IsLastDigitZero(ad) {
			break // Exit the loop when i is greater than or equal to 5
		}
		ad, err = transformer_utils.GenerateRandomInt(digitsAfterDecimal)

		if err != nil {
			return 0, fmt.Errorf("unable to generate a random int64 to convert to a float")
		}
	}

	if err != nil {
		return 0, fmt.Errorf("unable to generate a random after digits integer")
	}

	combinedStr := fmt.Sprintf("%d.%d", bd, ad)

	result, err := strconv.ParseFloat(combinedStr, 64)
	if err != nil {
		return 0, fmt.Errorf("unable to convert string to float")
	}

	returnValue = result

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
