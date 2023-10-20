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

	// int64 only supports 18 digits, so if the count => 19, this will error out
	if count >= 19 {
		return 0, fmt.Errorf("count has to be less than 18 digits since int64 only supports up to 18 digits")
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
		rand.Int generates a random number within the range [0, max-1], so if count == 8 [0 -> 9999999]. If the generated random integer is already the maximum possible value, then adding the minimum value to it will overflow it to count + 1. This is because the big.Int.Add() function adds two big integers together and returns a new big integer. If the first digit is a 9 and it's already count long then adding the min will overflow. So we only add if the digit count is not count AND the first digit is not 9.

	*/

	if FirstDigitIsNine(randInt.Int64()) && GetIntLength(randInt.Int64()) == count {
		return randInt.Int64(), nil
	} else {
		randInt.Add(randInt, minValue)
		return randInt.Int64(), nil

	}
}

func FirstDigitIsNine(n int64) bool {
	// Convert the int64 to a string
	str := strconv.FormatInt(n, 10)

	// Check if the string is empty or if the first character is '9'
	if len(str) > 0 && str[0] == '9' {
		return true
	}

	return false
}
