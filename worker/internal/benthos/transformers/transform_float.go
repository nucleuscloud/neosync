package transformers

import (
	"fmt"
	"time"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/nucleuscloud/neosync/worker/internal/rng"
)

func init() {
	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(bloblang.NewFloat64Param("randomization_range_min")).
		Param(bloblang.NewFloat64Param("randomization_range_max")).
		Param(bloblang.NewInt64Param("precision").Optional()).
		Param(bloblang.NewInt64Param("scale").Optional()).
		Param(bloblang.NewInt64Param("seed").Default(time.Now().UnixNano()))

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

		precision, err := args.GetOptionalInt64("precision")
		if err != nil {
			return nil, err
		}
		scale, err := args.GetOptionalInt64("scale")
		if err != nil {
			return nil, err
		}
		seed, err := args.GetInt64("seed")
		if err != nil {
			return nil, err
		}
		randomizer := rng.New(seed)

		return func() (any, error) {
			res, err := transformFloat(randomizer, valuePtr, rMin, rMax, precision, scale)
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

func transformFloat(randomizer rng.Rand, value *float64, rMin, rMax float64, precision, scale *int64) (*float64, error) {
	if value == nil {
		return nil, nil
	}

	minRange := *value - rMin
	maxRange := *value + rMax

	newVal, err := generateRandomFloat64(randomizer, false, minRange, maxRange, precision, scale)
	if err != nil {
		return nil, fmt.Errorf("unable to generate a random float64 with inclusive bounds with length [%f:%f]: %w", minRange, maxRange, err)
	}
	return &newVal, nil
}
