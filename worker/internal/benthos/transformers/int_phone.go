package neosync_transformers

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	"github.com/bxcodec/faker/v4"
)

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewBoolParam("preserve_length"))

	// register the plugin
	err := bloblang.RegisterMethodV2("intphonetransformer", spec, func(args *bloblang.ParsedParams) (bloblang.Method, error) {

		preserveLength, err := args.GetBool("preserve_length")
		if err != nil {
			return nil, err
		}

		return bloblang.Int64Method(func(s int64) (any, error) {
			res, err := ProcessIntPhoneNumber(s, preserveLength)
			return res, err
		}), nil
	})

	if err != nil {
		panic(err)
	}

}

// main transformer logic goes here
func ProcessIntPhoneNumber(pn int64, preserveLength bool) (int64, error) {

	var returnValue int64

	if preserveLength {

		numStr := strconv.FormatInt(pn, 10)

		val, err := GenerateRandomInt(int64(len(numStr))) // generates len(pn) random numbers from 0 -> 9

		if err != nil {
			return 0, fmt.Errorf("unable to generate phone number")
		}

		returnValue = val

	} else {

		str := strings.ReplaceAll(faker.Phonenumber(), "-", "")

		returnValue, err := strconv.ParseInt(str, 10, 64)

		if err != nil {
			return 0, fmt.Errorf("unable to convert phone number string to int64")
		}

		return returnValue, nil

	}

	return returnValue, nil
}

func GenerateRandomInt(count int64) (int64, error) {
	if count <= 0 {
		return 0, fmt.Errorf("count is zero or not a positive integer")
	}

	// Calculate the min and max values for count
	minValue := new(big.Int).Exp(big.NewInt(10), big.NewInt(count-1), nil)
	maxValue := new(big.Int).Exp(big.NewInt(10), big.NewInt(count), nil)

	// Generate a random integer within the specified range
	randInt, err := rand.Int(rand.Reader, maxValue)
	if err != nil {
		return 0, fmt.Errorf("unable to generate a random integer")
	}

	/*
	 rand.Int generates a random number within the range [0, max-1], where max is the upper bound of the range.  If we set count to 9, the upper bound maxVal will be the maximum 9-digit number (999999999), but rand.Int can generatee a random number up to maxVal-1 (i.e., 999999998), which can result in an 8-digit number. If the generated random integer is already the maximum possible value, then adding the minimum value to it will result in a new big integer with the desired number of digits. This is because the big.Int.Add() function adds two big integers together and returns a new big integer.

	*/

	// Add the minimum value to ensure it has the desired number of digits
	randInt.Add(randInt, minValue)

	// Convert the big integer to an int64 value
	randInt64 := randInt.Int64()

	return randInt64, nil
}
