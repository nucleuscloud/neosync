package transformers

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewInt64Param("int_length")).
		Param(bloblang.NewStringParam("sign"))

	err := bloblang.RegisterFunctionV2("generate_random_int", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		intLength, err := args.GetInt64("int_length")
		if err != nil {
			return nil, err
		}

		sign, err := args.GetString("sign")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {
			res, err := GenerateRandomInt(intLength, sign)
			return res, err
		}, nil
	})

	if err != nil {
		panic(err)
	}

}

// Generates a random integer up to 18 digits in length
// The sign param determines either a positive, negative or randomly assigned sign
func GenerateRandomInt(intLength int64, sign string) (int64, error) {

	var returnValue int64

	if sign != "positive" && sign != "negative" && sign != "random" {
		return 0, errors.New("sign can only be 'positive', 'negative', or 'random'")
	}

	if intLength <= 1 {
		return 0, errors.New("the length of the integer cannot be zero or less than zero")
	}

	if intLength > 10 {
		return 0, errors.New("the length of the integer cannot be greater than 18 digits")
	}

	// user didnt set the int length so we can randomly generate one of the default length four

	// max length of intLength is 9 digits so no issues with casting an int64 to an int32
	val, err := transformer_utils.GenerateRandomInt(int(intLength))
	if err != nil {
		return 0, fmt.Errorf("unable to generate a random string with length")
	}

	returnValue = int64(val)

	if sign == "random" {

		randInt := rand.Intn(2)
		if randInt == 1 {
			// return the positive value
			return returnValue, nil
		} else {
			// return the negative value
			return returnValue * -1, nil
		}

	} else if sign == "negative" {
		return returnValue * -1, nil
	} else {
		return returnValue, nil
	}

}

func IsNegativeInt(val int64) bool {
	if (val * -1) < 0 {
		return false
	} else {
		return true
	}
}
