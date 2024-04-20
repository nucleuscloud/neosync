package transformers

import (
	"fmt"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

func init() {
	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(bloblang.NewFloat64Param("randomization_range_min")).
		Param(bloblang.NewFloat64Param("randomization_range_max"))

	err := bloblang.RegisterFunctionV2("transform_float64", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		valuePtr, err := args.GetOptionalFloat64("value")
		if err != nil {
			return nil, err
		}

		rMin, err := args.GetFloat64("randomization_range_min")
		if err != nil {
			return nil, err
		}

		rMax, err := args.GetFloat64("randomization_range_max")
		if err != nil {
			return nil, err
		}
		return func() (any, error) {
			res, err := transformFloat(valuePtr, rMin, rMax)
			if err != nil {
				return nil, fmt.Errorf("unable to run transform_float64: %w", err)
			}
			return res, nil
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

func transformFloat(value *float64, rMin, rMax float64) (*float64, error) {
	if value == nil {
		return nil, nil
	}

	minRange := *value - rMin
	maxRange := *value + rMax

	val, err := transformer_utils.GenerateRandomFloat64WithInclusiveBounds(minRange, maxRange)
	if err != nil {
		return nil, fmt.Errorf("unable to generate a random float64 with inclusive bounds with length [%f:%f]: %w", minRange, maxRange, err)
	}
	return &val, nil
}
