package transformers

import (
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
)

// +neosyncTransformerBuilder:generate:generateInt64

/*
Integers can either be either put into non-numeric data types (varchar, text, integer) or numeric data types
such as numeric, float, double, etc.

Only numeric data types return a valid numeric_precision, non-numeric data types do not. Integer types take up
a fixed amount of memory, defined below:

SMALLINT: 2 bytes, range of -32,768 to 32,767
INTEGER: 4 bytes, range of -2,147,483,648 to 2,147,483,647
BIGINT: 8 bytes, range of -9,223,372,036,854,775,808 to 9,223,372,036,854,775,807

As a result they don't return a precision.

So we will need to understand what type of column the user is trying to insert the data into to figure out how to get the "size" of the column. Also, maxCharacterLength and NumericPrecision will never be a non-zero value at the same time.
*/

func init() {
	spec := bloblang.NewPluginSpec().
		Description("Generates a random int64 value with a default length of 4.").
		Category("int64").
		Param(bloblang.NewBoolParam("randomize_sign").Default(false).Description("A boolean indicating whether the sign of the float should be randomized.")).
		Param(bloblang.NewInt64Param("min").Default(1).Description("Specifies the minimum value for the generated int.")).
		Param(bloblang.NewInt64Param("max").Default(10000).Description("Specifies the maximum value for the generated int.")).
		Param(bloblang.NewInt64Param("seed").Optional().Description("An optional seed value used to generate deterministic outputs."))

	err := bloblang.RegisterFunctionV2("generate_int64", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		randomizeSign, err := args.GetBool("randomize_sign")
		if err != nil {
			return nil, err
		}

		min, err := args.GetInt64("min")
		if err != nil {
			return nil, err
		}

		max, err := args.GetInt64("max")
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
			res, err := generateRandomInt64(randomizer, randomizeSign, min, max)
			if err != nil {
				return nil, fmt.Errorf("unable to run generate_int64: %w", err)
			}
			return res, nil
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

func NewGenerateInt64OptsFromConfig(config *mgmtv1alpha1.GenerateInt64) (*GenerateInt64Opts, error) {
	if config == nil {
		return NewGenerateInt64Opts(
			nil,
			nil,
			nil,
			nil,
		)
	}
	return NewGenerateInt64Opts(
		config.RandomizeSign,
		config.Min,
		config.Max, nil,
	)
}

func (t *GenerateInt64) Generate(opts any) (any, error) {
	parsedOpts, ok := opts.(*GenerateInt64Opts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
	}

	return generateRandomInt64(parsedOpts.randomizer, parsedOpts.randomizeSign, parsedOpts.min, parsedOpts.max)
}

/*
Generates a random int64 in the interval [min, max].
*/
func generateRandomInt64(randomizer rng.Rand, randomizeSign bool, minValue, maxValue int64) (int64, error) {
	output, err := transformer_utils.GenerateRandomInt64InValueRange(randomizer, minValue, maxValue)
	if err != nil {
		return 0, err
	}
	if randomizeSign && generateRandomBool(randomizer) {
		output *= -1
	}
	return output, nil
}
