package neosync_transformers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
)

const defaultFloatLen = 5
const defaultFloatDecimals = 3

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
func ProcessRandomFloat(i float64, preserveLength bool, digitsBeforeDecimal, digitsAfterDecimal int64) (int64, error) {

	var returnValue float64

	fLen := GetFloatLength(i)

	if preserveLength {

		beforeDecimal, err := GenerateRandomIntWithLength(int64(GetIntLength(int64(fLen.DigitsBeforeDecimalLength))))

		afterDecimal, err := GenerateRandomIntWithLength(int64(fLen.DigitsAfterDecimalLength))

		if err != nil {
			return 0, fmt.Errorf("unable to generate a random string with length")
		}

		returnValue = 

	} else if intLength > 0 {

		val, err := GenerateRandomIntWithLength(intLength)

		if err != nil {
			return 0, fmt.Errorf("unable to generate a random string with length")
		}

		returnValue = val

	} else if preserveLength && intLength > 0 {

		val, err := GenerateRandomIntWithLength(intLength)

		if err != nil {
			return 0, fmt.Errorf("unable to generate a random string with length")
		}

		returnValue = val

	} else {

		val, err := GenerateRandomIntWithLength(defaultIntLength)

		if err != nil {
			return 0, fmt.Errorf("unable to generate a random string with length")
		}

		returnValue = val

	}

	return returnValue, nil
}

type FloatLength struct {
	DigitsBeforeDecimalLength int
	DigitsAfterDecimalLength  int
}

func GetFloatLength(i float64) *FloatLength {
	// Convert the int64 to a string
	str := strconv.FormatFloat(i, 'f', -1, 64)

	parsed := strings.Split(str, ".")

	return &FloatLength{
		DigitsBeforeDecimalLength: len(parsed[0]),
		DigitsAfterDecimalLength:  len(parsed[1]),
	}
}
