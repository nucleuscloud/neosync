package transformers

import (
	"fmt"
	"reflect"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
)

// +neosyncTransformerBuilder:transform:transformInt64

func init() {
	spec := bloblang.NewPluginSpec().
		Description("Anonymizes and transforms an existing int64 value.").
		Category("int64").
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(bloblang.NewInt64Param("randomization_range_min").Default(1).Description("Specifies the minimum value for the range of the int.")).
		Param(bloblang.NewInt64Param("randomization_range_max").Default(10000).Description("Specifies the maximum value for the range of the int.")).
		Param(bloblang.NewInt64Param("seed").Optional().Description("An optional seed value used to generate deterministic outputs."))

	err := bloblang.RegisterFunctionV2(
		"transform_int64",
		spec,
		func(args *bloblang.ParsedParams) (bloblang.Function, error) {
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

			seedArg, err := args.GetOptionalInt64("seed")
			if err != nil {
				return nil, err
			}

			seed, err := transformer_utils.GetSeedOrDefault(seedArg)
			if err != nil {
				return nil, err
			}

			randomizer := rng.New(seed)

			return func() (any, error) {
				res, err := transformInt(randomizer, valuePtr, rMin, rMax)
				if err != nil {
					return nil, fmt.Errorf("unable to run transform_int64: %w", err)
				}
				return res, nil
			}, nil
		},
	)

	if err != nil {
		panic(err)
	}
}

func NewTransformInt64OptsFromConfig(
	config *mgmtv1alpha1.TransformInt64,
) (*TransformInt64Opts, error) {
	if config == nil {
		return NewTransformInt64Opts(nil, nil, nil)
	}
	return NewTransformInt64Opts(
		config.RandomizationRangeMin,
		config.RandomizationRangeMax,
		nil,
	)
}

func (t *TransformInt64) Transform(value, opts any) (any, error) {
	parsedOpts, ok := opts.(*TransformInt64Opts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
	}

	return transformInt(
		parsedOpts.randomizer,
		value,
		parsedOpts.randomizationRangeMin,
		parsedOpts.randomizationRangeMax,
	)
}

func transformInt(randomizer rng.Rand, value any, rMin, rMax int64) (*int64, error) {
	if value == nil {
		return nil, nil
	}

	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, nil
		}
	}

	valueInt, err := transformer_utils.AnyToInt64(value)
	if err != nil {
		return nil, err
	}

	minRange := valueInt - rMin
	maxRange := valueInt + rMax

	val, err := transformer_utils.GenerateRandomInt64InValueRange(randomizer, minRange, maxRange)
	if err != nil {
		return nil, fmt.Errorf(
			"unable to generate a random int64 with length [%d:%d]:%w",
			minRange,
			maxRange,
			err,
		)
	}
	return &val, nil
}
