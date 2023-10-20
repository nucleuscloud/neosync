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

		val, err := GenerateRandomInt(len(numStr)) // generates len(pn) random numbers from 0 -> 9

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

func GenerateRandomInt(count int) (int64, error) {
	if count <= 0 {
		return 0, fmt.Errorf("count is zero or not a positive integer")
	}

	// Calculate the min and max vas
	minVal := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(count-1)), nil)
	maxVal := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(count)), nil)
	maxVal.Sub(maxVal, big.NewInt(1))

	// Generate a random integer within the specified range
	randInt, err := rand.Int(rand.Reader, maxVal)
	if err != nil {
		return 0, fmt.Errorf("unable to generate a random integer")
	}

	// Add the minimum value to ensure it has the desired number of digits
	randInt.Add(randInt, minVal)

	return randInt.Int64(), nil
}
