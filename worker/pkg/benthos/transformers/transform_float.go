package transformers

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"sync"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
)

// +neosyncTransformerBuilder:transform:transformFloat64

func init() {
	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(bloblang.NewFloat64Param("randomization_range_min")).
		Param(bloblang.NewFloat64Param("randomization_range_max")).
		Param(bloblang.NewInt64Param("precision").Optional()).
		Param(bloblang.NewInt64Param("scale").Optional()).
		Param(bloblang.NewInt64Param("seed").Optional())

	err := bloblang.RegisterFunctionV2("transform_float64", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
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
		var seed int64
		if seedArg != nil {
			seed = *seedArg
		} else {
			// we want a bit more randomness here with generate_email so using something that isn't time based
			var err error
			seed, err = transformer_utils.GenerateCryptoSeed()
			if err != nil {
				return nil, err
			}
		}
		randomizer := rng.New(seed)

		maxnumgetter := newMaxNumCache()

		return func() (any, error) {
			res, err := transformFloat(randomizer, maxnumgetter, value, rMin, rMax, precision, scale)
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

func (t *TransformFloat64) Transform(value, opts any) (any, error) {
	parsedOpts, ok := opts.(*TransformFloat64Opts)
	if !ok {
		return nil, errors.New("invalid parse opts")
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

func transformFloat(randomizer rng.Rand, maxnumgetter maxNum, value any, rMin, rMax float64, precision, scale *int64) (*float64, error) {
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
		return nil, fmt.Errorf("unable to generate a random float64 with inclusive bounds with length [%f:%f]: %w", minValue, maxValue, err)
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
