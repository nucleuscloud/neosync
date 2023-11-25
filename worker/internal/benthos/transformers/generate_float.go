package transformers

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewStringParam("sign")).
		Param(bloblang.NewInt64Param("digits_before_decimal")).
		Param(bloblang.NewInt64Param("digits_after_decimal"))

	err := bloblang.RegisterFunctionV2("generate_random_float", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		sign, err := args.GetString("sign")
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

		return func() (any, error) {
			res, err := GenerateRandomFloat(sign, digitsBeforeDecimal, digitsAfterDecimal)
			return res, err

		}, nil
	})

	if err != nil {
		panic(err)
	}

}

var random = "random"
var negative = "negative"

func GenerateRandomFloat(sign string, digitsBeforeDecimal, digitsAfterDecimal int64) (float64, error) {

	var returnValue float64

	if sign != "positive" && sign != negative && sign != random {
		return 0, errors.New("sign can only be 'positive', 'negative', or 'random'")
	}

	if digitsAfterDecimal < 1 || digitsAfterDecimal > 9 {
		return 0, errors.New("the length of the digits after the decimal cannot be less than one or greater than 9")
	}

	if digitsBeforeDecimal < 0 || digitsBeforeDecimal > 9 {
		return 0, errors.New("the length of the digits after the decimal cannot be less than one or greater than 9")
	}

	res, err := GenerateRandomFloatWithLength(int(digitsBeforeDecimal), int(digitsAfterDecimal))
	if err != nil {
		return 0, err
	}

	returnValue = res

	if sign == random {
		//nolint:all
		randInt := rand.Intn(2)
		if randInt == 1 {
			// return the positive value
			return returnValue, nil
		} else {
			// return the negative value
			return returnValue * -1, nil
		}

	} else if sign == negative {
		return returnValue * -1, nil
	} else {
		return returnValue, nil
	}
}

func GenerateRandomFloatWithLength(digitsBeforeDecimal, digitsAfterDecimal int) (float64, error) {

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

func IsNegativeFloat(val float64) bool {
	if (val * -1) < 0 {
		return false
	} else {
		return true
	}
}
