package neosync_transformers

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
)

const defaultIntLength = 4

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewBoolParam("preserve_length")).
		Param(bloblang.NewInt64Param("int_length"))

	// register the plugin
	err := bloblang.RegisterMethodV2("randominttransformer", spec, func(args *bloblang.ParsedParams) (bloblang.Method, error) {

		preserveLength, err := args.GetBool("preserve_length")
		if err != nil {
			return nil, err
		}

		intLength, err := args.GetInt64("int_length")
		if err != nil {
			return nil, err
		}

		if err != nil {
			return nil, fmt.Errorf("unable to convert the string case to a defined enum value")
		}

		return bloblang.Int64Method(func(i int64) (any, error) {
			res, err := ProcessRandomInt(i, preserveLength, intLength)
			return res, err
		}), nil
	})

	if err != nil {
		panic(err)
	}

}

// main transformer logic goes here
func ProcessRandomInt(i int64, preserveLength bool, intLength int64) (int64, error) {
	var returnValue int64

	if preserveLength {

		val, err := GenerateRandomInt(int64(GetIntLength(i)))

		if err != nil {
			return 0, fmt.Errorf("unable to generate a random string with length")
		}

		returnValue = val

	} else if intLength > 0 {

		val, err := GenerateRandomInt(intLength)

		if err != nil {
			return 0, fmt.Errorf("unable to generate a random string with length")
		}

		returnValue = val

	} else if preserveLength && intLength > 0 {

		val, err := GenerateRandomInt(int64(GetIntLength(i)))

		if err != nil {
			return 0, fmt.Errorf("unable to generate a random string with length")
		}

		returnValue = val

	} else {

		val, err := GenerateRandomInt(defaultIntLength)

		if err != nil {
			return 0, fmt.Errorf("unable to generate a random string with length")
		}

		returnValue = val

	}

	return returnValue, nil
}

func GenerateRandomIntWithLength(l int64) (int64, error) {
	if l <= 0 {
		return 0, fmt.Errorf("the length cannot be zero or negative")
	}

	newInt := 10

	newExpInt64 := int64(newInt)

	// Calculate the max value for l
	maxValue := new(big.Int).Exp(big.NewInt(newExpInt64), big.NewInt(l), nil)

	// Generate a random int64 value within the range
	randomValue, err := rand.Int(rand.Reader, maxValue)
	if err != nil {
		return 0, err
	}

	return randomValue.Int64(), nil
}

func GetIntLength(i int64) int {
	// Convert the int64 to a string
	str := strconv.FormatInt(i, 10)

	length := len(str)

	return length
}
