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
		Param(bloblang.NewFloat64Param("max"))

	err := bloblang.RegisterFunctionV2("generate_float", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

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

		return func() (any, error) {
			res, err := GenerateRandomFloat(randomizeSign, min, max)
			return res, err

		}, nil
	})

	if err != nil {
		panic(err)
	}

}

func GenerateRandomFloat(randomizeSign bool, min, max float64) (float64, error) {

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
		if randInt == 1 {
			// return the positive value
			return returnValue, nil
		} else {
			// return the negative value
			return returnValue * -1, nil
		}

	} else {
		res, err := transformer_utils.GenerateRandomFloat64WithInclusiveBounds(min, max)
		if err != nil {
			return 0, err
		}
		returnValue = res
	}
	return returnValue, nil
}
