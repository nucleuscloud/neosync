package transformers

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

func init() {
	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewInt64Param("max_length").Default(10000)).
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(bloblang.NewBoolParam("preserve_length").Default(false)).
		Param(bloblang.NewInt64Param("seed").Default(time.Now().UnixNano()))

	err := bloblang.RegisterFunctionV2("transform_last_name", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
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

		maxLength, err := args.GetInt64("max_length")
		if err != nil {
			return nil, err
		}

		seed, err := args.GetInt64("seed")
		if err != nil {
			return nil, err
		}
		randomizer := rand.New(rand.NewSource(seed)) //nolint:gosec

		return func() (any, error) {
			res, err := transformLastName(randomizer, value, preserveLength, maxLength)
			if err != nil {
				return nil, fmt.Errorf("unable to run transform_last_name: %w", err)
			}
			return res, nil
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

// Generates a random last name which can be of either random length between [2,12] characters or as long as the input name
func transformLastName(randomizer *rand.Rand, name string, preserveLength bool, maxLength int64) (*string, error) {
	if name == "" {
		return nil, nil
	}

	maxValue := maxLength

	// unable to generate a random name of this fixed size
	// we may want to change this to just use the below algorithm and pad so that it is more unique
	// as with this algorithm, it will only ever use values from the underlying map that are that specific size
	if preserveLength {
		maxValue = int64(len(name))
		output, err := generateRandomLastName(randomizer, &maxValue, maxValue)
		if err == nil {
			return &output, nil
		}
	}

	output, err := generateRandomLastName(randomizer, nil, maxValue)
	if err != nil {
		return nil, err
	}

	// pad the string so that we can get the correct value
	if preserveLength && int64(len(output)) != maxValue {
		output += transformer_utils.GetRandomCharacterString(randomizer, maxValue-int64(len(output)))
	}
	return &output, nil
}
