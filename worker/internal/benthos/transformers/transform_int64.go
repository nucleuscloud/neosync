package transformers

import (
	"fmt"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

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

		var value int64
		if valuePtr != nil {
			value = *valuePtr
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
			res, err := TransformInt(value, rMin, rMax)
			return res, err
		}, nil
	})

	if err != nil {
		panic(err)
	}

}

func TransformInt(value, rMin, rMax int64) (*int64, error) {

	if value == 0 {
		return nil, nil
	}

	// require that the value is in the randomization range so that we can transform it
	// otherwise, should use the generate_int transformer

	if !transformer_utils.IsInt64InRandomizationRange(value, rMin, rMax) {
		zeroVal := int64(0)
		return &zeroVal, fmt.Errorf("the value is not the provided range")
	}

	minRange := value - rMin
	maxRange := value + rMax

	val, err := transformer_utils.GenerateRandomInt64InValueRange(minRange, maxRange)
	if err != nil {

		return nil, fmt.Errorf("unable to generate a random string with length")

	}

	return &val, nil
}
