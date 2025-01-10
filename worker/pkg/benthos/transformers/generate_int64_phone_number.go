package transformers

import (
	"fmt"
	"strconv"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	transformers_dataset "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/data-sets"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

// +neosyncTransformerBuilder:generate:generateInt64PhoneNumber

var defaultPhoneNumberLength = int64(10)

func init() {
	spec := bloblang.NewPluginSpec().Description("Generates a new int64 phone number with a default length of 10.").
		Param(bloblang.NewInt64Param("seed").Optional().Description("An optional seed value used to generate deterministic outputs."))

	err := bloblang.RegisterFunctionV2("generate_int64_phone_number", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
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
			res, err := generateRandomInt64PhoneNumber(randomizer)
			if err != nil {
				return nil, fmt.Errorf("unable to run generate_int64_phone_number: %w", err)
			}
			return res, nil
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

func NewGenerateInt64PhoneNumberOptsFromConfig(config *mgmtv1alpha1.GenerateInt64PhoneNumber) (*GenerateInt64PhoneNumberOpts, error) {
	return NewGenerateInt64PhoneNumberOpts(nil)
}

func (t *GenerateInt64PhoneNumber) Generate(opts any) (any, error) {
	parsedOpts, ok := opts.(*GenerateInt64PhoneNumberOpts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
	}
	return generateRandomInt64PhoneNumber(parsedOpts.randomizer)
}

/* Generates a random 10 digit phone number with a valid US area code and returns it as an int64. */
func generateRandomInt64PhoneNumber(randomizer rng.Rand) (int64, error) {
	// get a random area code from the areacodes data set
	randAreaCodeStr, err := transformer_utils.GetRandomValueFromSlice(randomizer, transformers_dataset.UsAreaCodes)
	if err != nil {
		return 0, err
	}

	randAreaCode, err := strconv.ParseInt(randAreaCodeStr, 10, 64)
	if err != nil {
		return 0, err
	}

	// generate the rest of the phone number
	pn, err := transformer_utils.GenerateRandomInt64FixedLength(randomizer, defaultPhoneNumberLength-3)
	if err != nil {
		return 0, err
	}

	return randAreaCode*1e7 + pn, nil
}
