package transformers

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformers_dataset "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/data-sets"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

func init() {
	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewInt64Param("max_length").Default(10000)).
		Param(bloblang.NewInt64Param("seed").Default(time.Now().UnixNano()))

	err := bloblang.RegisterFunctionV2("generate_first_name", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
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
			output, err := generateRandomFirstName(randomizer, maxLength)
			if err != nil {
				return nil, fmt.Errorf("unable to run generate_first_name: %w", err)
			}
			return output, nil
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

func generateRandomFirstName(randomizer *rand.Rand, maxLength int64) (string, error) {
	candidates := transformer_utils.GetSmallerOrEqualNumbers(transformers_dataset.FirstNameIndices, maxLength)
	if len(candidates) == 0 {
		return "", fmt.Errorf("unable to find first name smaller than requested max length: %d", maxLength)
	}
	randIdx := randomizer.Int63n(int64(len(candidates)))
	firstNames := transformers_dataset.FirstNameMap[candidates[randIdx]]
	return firstNames[randomizer.Intn(len(firstNames))], nil
}
