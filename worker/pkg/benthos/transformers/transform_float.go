package transformers

import (
	"fmt"
	"math"
	"strconv"
	"sync"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
)

// +neosyncTransformerBuilder:transform:transformFloat64

func init() {
	spec := bloblang.NewPluginSpec().
		Description("Anonymizes and transforms an existing float value.").
		Category("float64").
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(bloblang.NewFloat64Param("randomization_range_min").Default(1).Description("Specifies the minimum value for the range of the float.")).
		Param(bloblang.NewFloat64Param("randomization_range_max").Default(10000).Description("Specifies the maximum value for the randomization range of the float.")).
		Param(bloblang.NewInt64Param("precision").Optional().Description("An optional parameter that defines the number of significant digits for the float.")).
		Param(bloblang.NewInt64Param("scale").Optional().Description("An optional parameter that defines the number of decimal places for the float.")).
		Param(bloblang.NewInt64Param("seed").Optional().Description("An optional seed value used for generating deterministic transformations."))

	err := bloblang.RegisterFunctionV2(
		"transform_float64",
		spec,
		func(args *bloblang.ParsedParams) (bloblang.Function, error) {
			value, err := args.Get("value")
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
			seedArg, err := args.GetOptionalInt64("seed")
			if err != nil {
				return nil, err
			}

			seed, err := transformer_utils.GetSeedOrDefault(seedArg)
			if err != nil {
				return nil, err
			}
			randomizer := rng.New(seed)

			maxnumgetter := newMaxNumCache()

			return func() (any, error) {
				res, err := transformFloat(
					randomizer,
					maxnumgetter,
					value,
					rMin,
					rMax,
					precision,
					scale,
				)
				if err != nil {
					return nil, fmt.Errorf("unable to run transform_float64: %w", err)
				}
				return res, nil
			}, nil
		},
	)

	if err != nil {
		panic(err)
	}
}

func NewTransformFloat64OptsFromConfig(
	config *mgmtv1alpha1.TransformFloat64,
	scale, precision *int64,
) (*TransformFloat64Opts, error) {
	if config == nil {
		return NewTransformFloat64Opts(nil, nil, nil, nil, nil)
	}
	return NewTransformFloat64Opts(
		config.RandomizationRangeMin,
		config.RandomizationRangeMax,
		precision,
		scale,
		nil,
	)
}

func (t *TransformFloat64) Transform(value, opts any) (any, error) {
	parsedOpts, ok := opts.(*TransformFloat64Opts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
	}

	maxnumgetter := newMaxNumCache()

	return transformFloat(
		parsedOpts.randomizer,
		maxnumgetter,
		value,
		parsedOpts.randomizationRangeMin,
		parsedOpts.randomizationRangeMax,
		parsedOpts.precision,
		parsedOpts.scale,
	)
}

func transformFloat(
	randomizer rng.Rand,
	maxnumgetter maxNum,
	value any,
	rMin, rMax float64,
	precision, scale *int64,
) (*float64, error) {
	if value == nil {
		return nil, nil
	}

	parsedVal, err := transformer_utils.AnyToFloat64(value)
	if err != nil {
		return nil, err
	}

	minValue := parsedVal - rMin
	maxValue := parsedVal + rMax

	if precision != nil {
		var scaleVal *int
		if scale != nil {
			newVal := int(*scale)
			scaleVal = &newVal
		}
		curbedMaxNum, err := maxnumgetter.CalculateMaxNumber(int(*precision), scaleVal)
		if err != nil {
			return nil, fmt.Errorf("unable to compute max number for the given precision and scale")
		}
		maxValue = transformer_utils.Ceil(maxValue, curbedMaxNum)
	}

	newVal, err := generateRandomFloat64(randomizer, false, minValue, maxValue, precision, scale)
	if err != nil {
		return nil, fmt.Errorf(
			"unable to generate a random float64 with inclusive bounds with length [%f:%f]: %w",
			minValue,
			maxValue,
			err,
		)
	}
	return &newVal, nil
}

func newMaxNumCache() *maxNumCache {
	return &maxNumCache{
		cache: map[string]float64{},
		mu:    sync.RWMutex{},
	}
}

type maxNumCache struct {
	cache map[string]float64
	mu    sync.RWMutex
}

type maxNum interface {
	CalculateMaxNumber(precision int, scale *int) (float64, error)
}

func (m *maxNumCache) CalculateMaxNumber(precision int, scale *int) (float64, error) {
	if precision <= 0 {
		return 0, fmt.Errorf("invalid precision value")
	}

	// If scale is nil, default it to zero
	actualScale := 0
	if scale != nil {
		actualScale = *scale
	}

	m.mu.RLock()
	key := m.computeKey(precision, actualScale)
	cachedVal, ok := m.cache[key]
	m.mu.RUnlock()
	if ok {
		return cachedVal, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	maxAllowedValue, err := calculateMaxNumber(precision, &actualScale)
	if err != nil {
		return 0, err
	}

	m.cache[key] = maxAllowedValue
	return maxAllowedValue, nil
}

func (m *maxNumCache) computeKey(precision, scale int) string {
	return fmt.Sprintf("%d_%d", precision, scale)
}

func calculateMaxNumber(precision int, scale *int) (float64, error) {
	if precision <= 0 {
		return 0, fmt.Errorf("invalid precision value")
	}

	// If scale is nil, default it to zero
	actualScale := 0
	if scale != nil && *scale > 0 {
		actualScale = *scale
	}
	// Calculate the number of integer digits
	intDigits := precision - actualScale
	if intDigits <= 0 {
		return 0, fmt.Errorf("invalid precision and scale combination")
	}

	// Construct the maximum integer part
	maxIntPart := math.Pow(10, float64(intDigits)) - 1

	// Construct the maximum fractional part
	maxFracPart := ""
	if actualScale > 0 {
		maxFracPart = fmt.Sprintf(".%0*d", actualScale, int(math.Pow(10, float64(actualScale))-1))
	}

	// Combine integer and fractional parts into a float
	maxAllowedStr := fmt.Sprintf("%d%s", int(maxIntPart), maxFracPart)
	maxAllowedValue, err := strconv.ParseFloat(maxAllowedStr, 64)
	if err != nil {
		return 0, err
	}
	return maxAllowedValue, nil
}
