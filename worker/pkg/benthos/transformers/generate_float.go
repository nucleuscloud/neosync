package transformers

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
)

// +neosyncTransformerBuilder:generate:generateFloat64

func init() {
	spec := bloblang.NewPluginSpec().
		Description("Generates a random floating point number with a max precision of 17. Go float64 adheres to the IEEE 754 standard for double-precision floating-point numbers.").
		Category("float64").
		Param(bloblang.NewBoolParam("randomize_sign").Default(false).Description("A boolean indicating whether the sign of the float should be randomized.")).
		Param(bloblang.NewFloat64Param("min").Default(1).Description("Specifies the minimum value for the generated float.")).
		Param(bloblang.NewFloat64Param("max").Default(10000).Description("Specifies the maximum value for the generated float")).
		Param(bloblang.NewInt64Param("precision").Optional().Description("An optional parameter that defines the number of significant digits for the generated float.")).
		Param(bloblang.NewInt64Param("scale").Optional().Description("An optional parameter that defines the number of decimal places for the generated float.")).
		Param(bloblang.NewInt64Param("seed").Optional().Description("An optional seed value used to generate deterministic outputs."))

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

func NewGenerateFloat64OptsFromConfig(config *mgmtv1alpha1.GenerateFloat64, scale *int64) (*GenerateFloat64Opts, error) {
	if config == nil {
		return NewGenerateFloat64Opts(nil, nil, nil, nil, nil, nil)
	}
	return NewGenerateFloat64Opts(
		config.RandomizeSign,
		config.Min,
		config.Max,
		config.Precision,
		nil,
		nil,
	)
}

func (t *GenerateFloat64) Generate(opts any) (any, error) {
	parsedOpts, ok := opts.(*GenerateFloat64Opts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
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

	if randomizeSign && generateRandomBool(randomizer) {
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
