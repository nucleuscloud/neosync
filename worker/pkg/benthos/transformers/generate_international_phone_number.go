package transformers

import (
	"errors"
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
)

// +neosyncTransformerBuilder:generate:generateInternationalPhoneNumber

func init() {
	spec := bloblang.NewPluginSpec().
		Category("string").
		Description("Generates a new random international phone number including the + sign and no hyphens.").
		Param(bloblang.NewInt64Param("min").Default(9).Description("Specifies the minimum value for the generated phone number.")).
		Param(bloblang.NewInt64Param("max").Default(15).Description("Specifies the maximum value for the generated phone number.")).
		Param(bloblang.NewInt64Param("seed").Optional().Description("An optional seed value used to generate deterministic outputs."))

	err := bloblang.RegisterFunctionV2("generate_e164_phone_number", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
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
			res, err := generateInternationalPhoneNumber(randomizer, min, max)
			if err != nil {
				return nil, fmt.Errorf("unable to run generate_international_phone_number: %w", err)
			}
			return res, nil
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

func NewGenerateInternationalPhoneNumberOptsFromConfig(config *mgmtv1alpha1.GenerateE164PhoneNumber) (*GenerateInternationalPhoneNumberOpts, error) {
	if config == nil {
		return NewGenerateInternationalPhoneNumberOpts(
			nil,
			nil,
			nil,
		)
	}
	return NewGenerateInternationalPhoneNumberOpts(
		config.Min,
		config.Max,
		nil,
	)
}

func (t *GenerateInternationalPhoneNumber) Generate(opts any) (any, error) {
	parsedOpts, ok := opts.(*GenerateInternationalPhoneNumberOpts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
	}

	return generateInternationalPhoneNumber(parsedOpts.randomizer, parsedOpts.min, parsedOpts.max)
}

/*  Generates a random phone number in e164 format in the length interval [min, max] with the min length == 9 and the max length == 15.
 */
func generateInternationalPhoneNumber(randomizer rng.Rand, minValue, maxValue int64) (string, error) {
	if minValue < 9 || maxValue > 15 {
		return "", errors.New("the length has between 9 and 15 characters long")
	}

	val, err := transformer_utils.GenerateRandomInt64InLengthRange(randomizer, minValue, maxValue)
	if err != nil {
		return "", nil
	}

	return fmt.Sprintf("+%d", val), nil
}
