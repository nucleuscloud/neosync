package transformers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

var (
	defaultDigitsBeforeDecimal = 3
	defaultDigitsAfterDecimal  = 3
)

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(bloblang.NewBoolParam("preserve_length")).
		Param(bloblang.NewBoolParam("preserve_sign"))

	err := bloblang.RegisterFunctionV2("transform_float", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		valuePtr, err := args.GetOptionalFloat64("value")
		if err != nil {
			return nil, err
		}

		var value float64
		if valuePtr != nil {
			value = *valuePtr
		}

		preserveLength, err := args.GetBool("preserve_length")
		if err != nil {
			return nil, err
		}

		preserveSign, err := args.GetBool("preserve_sign")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {
			res, err := TransformFloat(value, preserveLength, preserveSign)
			return res, err

		}, nil
	})

	if err != nil {
		panic(err)
	}

}

func TransformFloat(value float64, preserveLength, preserveSign bool) (*float64, error) {

	var returnValue float64

	if value == 0 {
		return nil, nil
	}

	if preserveLength {
		fLen := GetFloatLength(value)
		res, err := GenerateRandomFloatWithDefinedLength(fLen.DigitsBeforeDecimalLength, fLen.DigitsAfterDecimalLength)

		if err != nil {
			return nil, err
		}

		returnValue = res
	} else {
		res, err := GenerateRandomFloatWithDefinedLength(defaultDigitsBeforeDecimal, defaultDigitsAfterDecimal)
		if err != nil {
			return nil, err
		}

		returnValue = res
	}

	if preserveSign {

		if value < 0 {
			res := returnValue * -1
			return &res, nil
		} else {
			return &returnValue, nil
		}

	} else {
		return &returnValue, nil
	}

}

func GenerateRandomFloatWithDefinedLength(digitsBeforeDecimal, digitsAfterDecimal int) (float64, error) {

	var returnValue float64

	bd, err := transformer_utils.GenerateRandomInt(digitsBeforeDecimal)
	if err != nil {
		return 0, fmt.Errorf("unable to generate a random before digits integer")
	}

	ad, err := transformer_utils.GenerateRandomInt(digitsAfterDecimal)

	// generate a new number if it ends in a zero so that the trailing zero doesn't get stripped and return
	// a value that is shorter than what the user asks for. This happens in when we convert the string to a float64
	for {
		if !transformer_utils.IsLastDigitZero(int64(ad)) {
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
	// Convert the float64 to a string
	str := fmt.Sprintf("%g", i)

	parsed := strings.Split(str, ".")

	return &FloatLength{
		DigitsBeforeDecimalLength: len(parsed[0]),
		DigitsAfterDecimalLength:  len(parsed[1]),
	}
}
