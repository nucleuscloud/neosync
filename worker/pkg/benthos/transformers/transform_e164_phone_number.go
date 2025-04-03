package transformers

import (
	"errors"
	"fmt"
	"strings"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
)

// +neosyncTransformerBuilder:transform:transformE164PhoneNumber

func init() {
	spec := bloblang.NewPluginSpec().
		Description("Transforms an existing E164 formatted phone number.").
		Category("string").
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(bloblang.NewBoolParam("preserve_length").Default(false).Description("Whether the original length of the input data should be preserved during transformation. If set to true, the transformation logic will ensure that the output data has the same length as the input data.")).
		Param(bloblang.NewInt64Param("max_length").Default(15).Description("Specifies the maximum length for the transformed data. This field ensures that the output does not exceed a certain number of characters.")).
		Param(bloblang.NewInt64Param("seed").Optional().Description("An optional seed value used to generate deterministic outputs."))

	err := bloblang.RegisterFunctionV2(
		"transform_e164_phone_number",
		spec,
		func(args *bloblang.ParsedParams) (bloblang.Function, error) {
			valuePtr, err := args.GetOptionalString("value")
			if err != nil {
				return nil, err
			}

			var value string
			if valuePtr != nil {
				value = *valuePtr
			}

			preserveLength, err := args.GetBool("preserve_length")
			if err != nil {
				return nil, err
			}

			maxLength, err := args.GetOptionalInt64("max_length")
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
				res, err := transformE164PhoneNumber(randomizer, value, preserveLength, maxLength)
				if err != nil {
					return nil, fmt.Errorf("unable to run transform_e164_phone_number: %w", err)
				}
				return res, nil
			}, nil
		},
	)

	if err != nil {
		panic(err)
	}
}

func NewTransformE164PhoneNumberOptsFromConfig(
	config *mgmtv1alpha1.TransformE164PhoneNumber,
	maxLength *int64,
) (*TransformE164PhoneNumberOpts, error) {
	if config == nil {
		return NewTransformE164PhoneNumberOpts(nil, nil, nil)
	}
	return NewTransformE164PhoneNumberOpts(config.PreserveLength, maxLength, nil)
}

func (t *TransformE164PhoneNumber) Transform(value, opts any) (any, error) {
	parsedOpts, ok := opts.(*TransformE164PhoneNumberOpts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
	}

	valueStr, ok := value.(string)
	if !ok {
		return nil, errors.New("value is not a string")
	}

	return transformE164PhoneNumber(
		parsedOpts.randomizer,
		valueStr,
		parsedOpts.preserveLength,
		&parsedOpts.maxLength,
	)
}

// Generates a random phone number and returns it as a string
func transformE164PhoneNumber(
	randomizer rng.Rand,
	phone string,
	preserveLength bool,
	maxLength *int64,
) (*string, error) {
	var returnValue string

	if phone == "" {
		return nil, nil
	}

	if preserveLength {
		res, err := generateE164FormatPhoneNumberPreserveLength(randomizer, phone)
		if err != nil {
			return nil, err
		}

		returnValue = res
	} else {
		minValue := int64(9)
		maxValue := int64(15)
		if maxLength != nil && *maxLength != 0 {
			maxValue = *maxLength
		}

		res, err := generateInternationalPhoneNumber(randomizer, minValue, maxValue)
		if err != nil {
			return nil, err
		}
		returnValue = res
	}

	return &returnValue, nil
}

// generates a random E164 phone number and returns it as a string
func generateE164FormatPhoneNumberPreserveLength(
	randomizer rng.Rand,
	number string,
) (string, error) {
	numberWithoutPlus := strings.TrimPrefix(number, "+")

	length := int64(len(numberWithoutPlus))

	vals, err := transformer_utils.GenerateRandomInt64FixedLength(randomizer, length)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("+%d", vals), nil
}
