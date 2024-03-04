package transformers

import (
	"math/rand"

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
			res, err := GenerateRandomInt64(randomizeSign, min, max)
			return res, err
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

/*
Generates a random int64 up to 18 digits in the interval [min, max].
*/
func GenerateRandomInt64(randomizeSign bool, min, max int64) (int64, error) {
	var returnValue int64

	if randomizeSign {
		res, err := transformer_utils.GenerateRandomInt64InValueRange(transformer_utils.AbsInt64(min), transformer_utils.AbsInt64(max))
		if err != nil {
			return 0, err
		}

		returnValue = res
		//nolint:all
		randInt := rand.Intn(2)
		if randInt == 1 {
			// return the positive value
			return returnValue, nil
		} else {
			// return the negative value
			return returnValue * -1, nil
		}
	} else {
		res, err := transformer_utils.GenerateRandomInt64InValueRange(min, max)
		if err != nil {
			return 0, err
		}

		returnValue = res
	}

	return returnValue, nil
}
