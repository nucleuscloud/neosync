package transformers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
)

// +neosyncTransformerBuilder:generate:generateInternationalPhoneNumber

func init() {
	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewInt64Param("min")).
		Param(bloblang.NewInt64Param("max"))

	err := bloblang.RegisterFunctionV2("generate_e164_phone_number", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		min, err := args.GetInt64("min")
		if err != nil {
			return nil, err
		}

		max, err := args.GetInt64("max")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {
			res, err := generateInternationalPhoneNumber(min, max)
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

func (t *GenerateInternationalPhoneNumber) Generate(opts any) (any, error) {
	parsedOpts, ok := opts.(*GenerateInternationalPhoneNumberOpts)
	if !ok {
		return nil, errors.New("invalid parse opts")
	}

	return generateInternationalPhoneNumber(parsedOpts.min, parsedOpts.max)
}

/*  Generates a random phone number in e164 format in the length interval [min, max] with the min length == 9 and the max length == 15.
 */
func generateInternationalPhoneNumber(minValue, maxValue int64) (string, error) {
	if minValue < 9 || maxValue > 15 {
		return "", errors.New("the length has between 9 and 15 characters long")
	}

	val, err := transformer_utils.GenerateRandomInt64InLengthRange(minValue, maxValue)
	if err != nil {
		return "", nil
	}

	return fmt.Sprintf("+%d", val), nil
}

func validateE164(p string) bool {
	if len(p) >= 10 && len(p) <= 15 && strings.Contains(p, "+") {
		return true
	}
	return false
}
