package transformers

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
)

// +neosyncTransformerBuilder:generate:generateFloat64

func init() {
	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewBoolParam("randomize_sign").Default(false)).
		Param(bloblang.NewFloat64Param("min")).
		Param(bloblang.NewFloat64Param("max")).
		Param(bloblang.NewInt64Param("precision").Optional()).
		Param(bloblang.NewInt64Param("scale").Optional()).
		Param(bloblang.NewInt64Param("seed").Default(time.Now().UnixNano()))

	err := bloblang.RegisterFunctionV2("generate_float64", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		randomizeSign, err := args.GetBool("randomize_sign")
		if err != nil {
			return nil, err
		}

		minVal, err := args.GetFloat64("min")
		if err != nil {
			return nil, err
		}

		maxVal, err := args.GetFloat64("max")
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
			res, err := generateRandomFloat64(randomizer, randomizeSign, minVal, maxVal, precision, scale)
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

func (t *GenerateFloat64) Generate(opts any) (any, error) {
	parsedOpts, ok := opts.(*GenerateFloat64Opts)
	if !ok {
		return nil, errors.New("invalid parse opts")
	}

	return generateRandomFloat64(
		parsedOpts.randomizer,
		parsedOpts.randomizeSign,
		parsedOpts.min,
		parsedOpts.max,
		parsedOpts.precision,
		parsedOpts.scale,
	)
}

/* Generates a random float64 value within the interval [min, max]*/
func generateRandomFloat64(
	randomizer rng.Rand,
	randomizeSign bool,
	minValue, maxValue float64,
	precision, scale *int64,
) (float64, error) {
	randomFloat, err := transformer_utils.GenerateRandomFloat64WithInclusiveBounds(randomizer, minValue, maxValue)
	if err != nil {
		return 0, err
	}

	// Apply scale if specified
	if scale != nil {
		randomFloat, err = roundToScale(randomFloat, int(*scale))
		if err != nil {
			return 0, err
		}
	}

	// Apply precision if specified
	if precision != nil {
		// Convert float to string to manipulate precision
		strFloat := fmt.Sprintf("%.*f", int(math.Max(0, float64(*precision))), randomFloat)

		// Trim or pad the number to match the exact precision if needed
		dotIndex := strings.Index(strFloat, ".")
		if dotIndex == -1 { // no decimal point, all digits are before the decimal
			dotIndex = len(strFloat) // treat end of string as start of non-existent fractional part
		}

		// Calculate digits before and after the decimal
		digitsBefore := dotIndex
		digitsAfter := len(strFloat) - dotIndex - 1

		if int64(digitsBefore+digitsAfter) > *precision { // total digits exceed precision
			if digitsBefore > int(*precision) { // need to cut in the integer part
				strFloat = strFloat[:int(*precision)]
				randomFloat, err = strconv.ParseFloat(strFloat, 64)
				if err != nil {
					return 0, err
				}
			} else { // cut in the fractional part
				allowedAfter := int(*precision) - digitsBefore
				strFloat = fmt.Sprintf("%.*f", allowedAfter, randomFloat)
				randomFloat, err = strconv.ParseFloat(strFloat, 64)
				if err != nil {
					return 0, err
				}
			}
		}
	}

	if randomizeSign && generateRandomizerBool(randomizer) {
		randomFloat *= -1
	}

	return randomFloat, nil
}

func roundToScale(val float64, scale int) (float64, error) {
	// Use strconv.FormatFloat to format and round to the desired scale
	formattedStr := strconv.FormatFloat(val, 'f', scale, 64)

	// Convert the formatted string back to float64
	roundedVal, err := strconv.ParseFloat(formattedStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to convert rounded string back to float64: %w", err)
	}

	return roundedVal, nil
}
