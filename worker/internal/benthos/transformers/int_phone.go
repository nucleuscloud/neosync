package neosync_transformers

import (
	"fmt"
	"math/rand"
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

		const maxPhoneNum = 9

		numStr := strconv.FormatInt(pn, 10)

		val, err := GenerateRandomInt(0, maxPhoneNum, len(numStr)) // generates len(pn) random numbers from 0 -> 9

		if err != nil {
			return 0, fmt.Errorf("unable to generate phone number")
		}

		for _, int := range val {
			returnValue = returnValue*10 + int64(int)
		}

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

func GenerateRandomInt(minInt, maxInt, count int) ([]int, error) {
	if count <= 0 {
		return nil, fmt.Errorf("count is zero or not an int")
	}

	randomInts := make([]int, count)
	for i := 0; i < count; i++ {
		randomInts[i] = rand.Intn(maxInt-minInt+1) + minInt
	}

	return randomInts, nil
}
