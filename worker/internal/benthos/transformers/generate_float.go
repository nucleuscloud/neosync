package transformers

import (
	"fmt"

	"github.com/benthosdev/benthos/v4/public/bloblang"
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
			res, err := generateRandomFloat64(randomizeSign, min, max, precision)
			if err != nil {
				return nil, fmt.Errorf("unable to run generate_float: %w", err)
			}
			return res, nil
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

/* Generates a random float64 value within the interval [min, max]*/
func generateRandomFloat64(randomizeSign bool, min, max float64, precision int64) (float64, error) {
	generatedVal, err := transformer_utils.GenerateRandomFloat64WithInclusiveBounds(min, max)
	if err != nil {
		return 0, err
	}

	if randomizeSign && generateRandomBool() {
		generatedVal *= -1.
	}

	reducedVal, err := transformer_utils.ReduceFloat64Precision(int(precision), generatedVal)
	if err != nil {
		return 0, err
	}
	return reducedVal, nil
}
