package neosync_benthos_transformers_methods

import (
	"fmt"
	"strconv"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
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
			res, err := GenerateIntPhoneNumber(s, preserveLength)
			return res, err
		}), nil
	})

	if err != nil {
		panic(err)
	}

}

// generates a random phone number and returns it as an int64
func GenerateIntPhoneNumber(number int64, preserveLength bool) (int64, error) {

	if preserveLength {

		res, err := GenerateIntPhoneNumberPreserveLength(number)
		if err != nil {
			return 0, fmt.Errorf("unable to convert phone number string to int64")
		}
		return res, err

	} else {

		res, err := GenerateRandomTenDigitIntPhoneNumber()
		if err != nil {
			return 0, fmt.Errorf("unable to convert phone number string to int64")
		}

		return res, err

	}

}

func GenerateIntPhoneNumberPreserveLength(number int64) (int64, error) {
	numStr := strconv.FormatInt(number, 10)

	val, err := transformer_utils.GenerateRandomInt(int64(len(numStr))) // generates len(pn) random numbers from 0 -> 9
	if err != nil {
		return 0, fmt.Errorf("unable to generate phone number")
	}

	return val, err

}

func GenerateRandomTenDigitIntPhoneNumber() (int64, error) {

	res, err := transformer_utils.GenerateRandomInt(int64(10))

	if err != nil {
		return 0, fmt.Errorf("unable to generate random phone number")
	}

	return res, nil
}
