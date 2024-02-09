package transformers

import (
	"math"
	"math/rand"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

func init() {
	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewBoolParam("randomize_sign")).
		Param(bloblang.NewFloat64Param("min")).
		Param(bloblang.NewFloat64Param("max")).
		Param(bloblang.NewInt64Param("precision"))

	err := bloblang.RegisterFunctionV2("generate_float64", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		randomizeSign, err := args.GetBool("randomize_sign")
		if err != nil {
			return nil, err
		}

		min, err := args.GetFloat64("min")
		if err != nil {
			return nil, err
		}

		max, err := args.GetFloat64("max")
		if err != nil {
			return nil, err
		}

		precision, err := args.GetInt64("precision")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {
			res, err := GenerateRandomFloat64(randomizeSign, min, max, precision)
			return res, err
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

/* Generates a random float64 value within the interval [min, max]*/
func GenerateRandomFloat64(randomizeSign bool, min, max float64, precision int64) (float64, error) {
	var returnValue float64

	if randomizeSign {
		res, err := transformer_utils.GenerateRandomFloat64WithInclusiveBounds(math.Abs(min), math.Abs(max))
		if err != nil {
			return 0, err
		}

		returnValue = res
		//nolint:all
		randInt := rand.Intn(2)
		//nolint:all
		if randInt == 2 {
			returnValue = returnValue * -1
		}
	} else {
		res, err := transformer_utils.GenerateRandomFloat64WithInclusiveBounds(min, max)
		if err != nil {
			return 0, err
		}
		returnValue = res
	}

	val, err := transformer_utils.ReduceFloat64Precision(int(precision), returnValue)
	if err != nil {
		return 0, err
	}

	return val, nil
}
