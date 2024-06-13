package transformers

import (
	"fmt"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformers_dataset "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/data-sets"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
)

var defaultPhoneNumberLength = int64(10)

func init() {
	spec := bloblang.NewPluginSpec()

	err := bloblang.RegisterFunctionV2("generate_int64_phone_number", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		return func() (any, error) {
			res, err := generateRandomInt64PhoneNumber()
			if err != nil {
				return nil, fmt.Errorf("unable to run generate_int64_phone_number: %w", err)
			}
			return res, nil
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

/* Generates a random 10 digit phone number with a valid US area code and returns it as an int64. */
func generateRandomInt64PhoneNumber() (int64, error) {
	ac := transformers_dataset.AreaCodes

	// get a random area code from the areacodes data set
	randAreaCode, err := transformer_utils.GetRandomValueFromSlice[int64](ac)
	if err != nil {
		return 0, err
	}

	// generate the rest of the phone number
	pn, err := transformer_utils.GenerateRandomInt64FixedLength(defaultPhoneNumberLength - 3)
	if err != nil {
		return 0, err
	}

	return randAreaCode*1e7 + pn, nil
}
