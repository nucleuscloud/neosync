package transformers

import (
	"math/rand"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

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
