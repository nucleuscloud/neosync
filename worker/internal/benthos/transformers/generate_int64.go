package transformers

import (
	"fmt"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

/*
Integers can either be either put into non-numeric data types (varchar, text, integer) or numeric data types
such as numeric, float, double, etc.

Only numeric data types return a valid numeric_precision, non-numeric data types do not. Integer types take up
a fixed amount of memory, defined below:

SMALLINT: 2 bytes, range of -32,768 to 32,767
INTEGER: 4 bytes, range of -2,147,483,648 to 2,147,483,647
BIGINT: 8 bytes, range of -9,223,372,036,854,775,808 to 9,223,372,036,854,775,807

As a result they don't return a precision.

So we will need to understand what type of column the user is trying to insert the data into to figure out how to get the "size" of the column. Also, maxCharacterLength and NumericPrecision will never be a non-zero value at the same time.
*/

func init() {
	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewBoolParam("randomize_sign")).
		Param(bloblang.NewInt64Param("min")).
		Param(bloblang.NewInt64Param("max"))

	err := bloblang.RegisterFunctionV2("generate_int64", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		randomizeSign, err := args.GetBool("randomize_sign")
		if err != nil {
			return nil, err
		}

		min, err := args.GetInt64("min")
		if err != nil {
			return nil, err
		}

		max, err := args.GetInt64("max")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {
			res, err := generateRandomInt64(randomizeSign, min, max)
			if err != nil {
				return nil, fmt.Errorf("unable to run generate_int64: %w", err)
			}
			return res, nil
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

/*
Generates a random int64 in the interval [min, max].
*/
func generateRandomInt64(randomizeSign bool, min, max int64) (int64, error) {
	output, err := transformer_utils.GenerateRandomInt64InValueRange(min, max)
	if err != nil {
		return 0, err
	}
	if randomizeSign && generateRandomBool() {
		output *= -1
	}
	return output, nil
}
