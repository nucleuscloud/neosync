package transformers

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformers_dataset "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/data-sets"
)

func init() {
	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewInt64Param("max_length").Default(10000)).
		Param(bloblang.NewInt64Param("seed").Default(time.Now().UnixNano()))

	err := bloblang.RegisterFunctionV2("generate_last_name", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
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
			output, err := generateRandomLastName(randomizer, nil, maxLength)
			if err != nil {
				return nil, fmt.Errorf("unable to run generate_last_name")
			}
			return output, nil
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

func generateRandomLastName(randomizer *rand.Rand, minLength *int64, maxLength int64) (string, error) {
	candidates := getFilteredNumbers(transformers_dataset.LastNameIndices, minLength, &maxLength)
	if len(candidates) == 0 {
		return "", fmt.Errorf("unable to find last name with range %s", getRangeText(minLength, maxLength))
	}
	randIdx := randomizer.Int63n(int64(len(candidates)))
	lastNames := transformers_dataset.LastNameMap[candidates[randIdx]]
	return lastNames[randomizer.Intn(len(lastNames))], nil
}
