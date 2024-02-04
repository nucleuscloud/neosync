package transformers

import (
	"fmt"
	"math/rand"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
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
		Param(bloblang.NewInt64Param("max")).
		Param(bloblang.NewInt64Param("numeric_precision")).
		Param(bloblang.NewInt64Param("max_length"))

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
		numericPrecision, err := args.GetInt64("numeric_precision")
		if err != nil {
			return nil, err
		}

		maxLength, err := args.GetInt64("max_length")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {
			res, err := GenerateRandomInt64(randomizeSign, min, max, numericPrecision, maxLength)
			return res, err
		}, nil
	})

	if err != nil {
		panic(err)
	}

}

// TODO: we might not want to do this with integers actually. Changing the length adn value of integers fundamentally changes the meaning of the value and I think should be up to the user to ensure they are doing it correclty. We shouldn't try out best guess to make this fit. We surface the rror and let them resolve it. For example, say we have:  min := int64(10), max := int64(2000) and the maxLength or numericPrecision is 3. And we generate the number 1999, we chop it to down to 999? or 199? there is a big difference between those two numbres and the user should resolve that IMO.
/*
Generates a random int64 up to 18 digits in the interval [min, max].
*/
func GenerateRandomInt64(randomizeSign bool, min, max, numericPrecision, maxLength int64) (int64, error) {

	var returnValue int64

	// maxLength will not be 0 if the column is a non-numeric data type
	if maxLength != 0 && transformer_utils.GetInt64Length(max) > maxLength {

		if transformer_utils.GetInt64Length(min) > maxLength {
			fmt.Println("hit min > maxlength")
			res, err := transformer_utils.GenerateRandomInt64InValueRange(transformer_utils.AbsInt64(min), transformer_utils.AbsInt64(max))
			if err != nil {
				return 0, err
			}

			returnValue = res[:maxLength]
		} else {
			fmt.Println("hit min < maxLength")

			res, err := transformer_utils.GenerateRandomInt64InValueRange(transformer_utils.AbsInt64(min), transformer_utils.AbsInt64(maxLength))
			if err != nil {
				return 0, err
			}

			returnValue = res

		}

	} else {
		// column is a numeric data-type
		if numericPrecision < transformer_utils.GetInt64Length(max) {

			res, err := transformer_utils.GenerateRandomInt64InValueRange(min, numericPrecision)
			if err != nil {
				return 0, err
			}

			returnValue = res
		} else {

			res, err := transformer_utils.GenerateRandomInt64InValueRange(min, max)
			if err != nil {
				return 0, err
			}

			returnValue = res

		}

	}

	if randomizeSign && rand.Intn(2) == 0 {
		return -returnValue, nil
	} else {
		return returnValue, nil
	}

}
