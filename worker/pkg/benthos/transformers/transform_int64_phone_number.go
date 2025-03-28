package transformers

import (
	"fmt"
	"reflect"
	"strconv"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	transformers_dataset "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/data-sets"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
)

// +neosyncTransformerBuilder:transform:transformInt64PhoneNumber

func init() {
	spec := bloblang.NewPluginSpec().
		Description("Anonymizes and transforms an existing int64 phone number.").
		Category("int64").
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(bloblang.NewBoolParam("preserve_length").Default(false).Description("Whether the original length of the input data should be preserved during transformation. If set to true, the transformation logic will ensure that the output data has the same length as the input data.")).
		Param(bloblang.NewInt64Param("seed").Optional().Description("An optional seed value used to generate deterministic outputs."))

	err := bloblang.RegisterFunctionV2(
		"transform_int64_phone_number",
		spec,
		func(args *bloblang.ParsedParams) (bloblang.Function, error) {
			valuePtr, err := args.GetOptionalInt64("value")
			if err != nil {
				return nil, err
			}

			var value int64
			if valuePtr != nil {
				value = *valuePtr
			}

			preserveLength, err := args.GetBool("preserve_length")
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
				res, err := transformInt64PhoneNumber(randomizer, value, preserveLength)
				if err != nil {
					return nil, fmt.Errorf("unable to run transform_int64_phone_number: %w", err)
				}
				return res, nil
			}, nil
		},
	)

	if err != nil {
		panic(err)
	}
}

func NewTransformInt64PhoneNumberOptsFromConfig(
	config *mgmtv1alpha1.TransformInt64PhoneNumber,
) (*TransformInt64PhoneNumberOpts, error) {
	if config == nil {
		return NewTransformInt64PhoneNumberOpts(nil, nil)
	}
	return NewTransformInt64PhoneNumberOpts(config.PreserveLength, nil)
}

func (t *TransformInt64PhoneNumber) Transform(value, opts any) (any, error) {
	parsedOpts, ok := opts.(*TransformInt64PhoneNumberOpts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
	}

	return transformInt64PhoneNumber(parsedOpts.randomizer, value, parsedOpts.preserveLength)
}

// generates a random phone number and returns it as an int64
func transformInt64PhoneNumber(
	randomizer rng.Rand,
	value any,
	preserveLength bool,
) (*int64, error) {
	if value == nil {
		return nil, nil
	}

	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, nil
		}
	}

	number, err := transformer_utils.AnyToInt64(value)
	if err != nil {
		return nil, err
	}

	if number == 0 {
		return nil, nil
	}

	if preserveLength {
		res, err := generateIntPhoneNumberPreserveLength(randomizer, number)
		if err != nil {
			return nil, err
		}
		return &res, nil
	} else {
		res, err := generateRandomInt64PhoneNumber(randomizer)
		if err != nil {
			return nil, err
		}
		return &res, nil
	}
}

func generateIntPhoneNumberPreserveLength(randomizer rng.Rand, number int64) (int64, error) {
	// get a random area code from the areacodes data set
	randAreaCodeStr, err := transformer_utils.GetRandomValueFromSlice(
		randomizer,
		transformers_dataset.UsAreaCodes,
	)
	if err != nil {
		return 0, err
	}

	randAreaCode, err := strconv.ParseInt(randAreaCodeStr, 10, 64)
	if err != nil {
		return 0, err
	}

	pn, err := transformer_utils.GenerateRandomInt64FixedLength(
		randomizer,
		transformer_utils.GetInt64Length(number)-3,
	)
	if err != nil {
		return 0, err
	}

	return randAreaCode*1e7 + pn, nil
}
