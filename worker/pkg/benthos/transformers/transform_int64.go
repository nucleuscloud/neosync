package transformers

import (
	"errors"
	"fmt"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
)

// +neosyncTransformerBuilder:transform:transformInt64

func init() {
	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(bloblang.NewInt64Param("randomization_range_min")).
		Param(bloblang.NewInt64Param("randomization_range_max"))

	err := bloblang.RegisterFunctionV2("transform_int64", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		valuePtr, err := args.GetOptionalInt64("value")
		if err != nil {
			return nil, err
		}

		rMin, err := args.GetInt64("randomization_range_min")
		if err != nil {
			return nil, err
		}

		rMax, err := args.GetInt64("randomization_range_max")
		if err != nil {
			return nil, err
		}
		return func() (any, error) {
			res, err := transformInt(valuePtr, rMin, rMax)
			if err != nil {
				return nil, fmt.Errorf("unable to run transform_int64: %w", err)
			}
			return res, nil
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

func (t *TransformInt64) Transform(value, opts any) (any, error) {
	parsedOpts, ok := opts.(*TransformInt64Opts)
	if !ok {
		return nil, errors.New("invalid parse opts")
	}

	valueInt, ok := value.(int64)
	if !ok {
		return nil, errors.New("value is not a string")
	}

	return transformInt(&valueInt, parsedOpts.randomizationRangeMin, parsedOpts.randomizationRangeMax)
}

func transformInt(value *int64, rMin, rMax int64) (*int64, error) {
	if value == nil {
		return nil, nil
	}

	minRange := *value - rMin
	maxRange := *value + rMax

	val, err := transformer_utils.GenerateRandomInt64InValueRange(minRange, maxRange)
	if err != nil {
		return nil, fmt.Errorf("unable to generate a random int64 with length [%d:%d]:%w", minRange, maxRange, err)
	}
	return &val, nil
}
