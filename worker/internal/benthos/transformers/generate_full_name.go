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

	err := bloblang.RegisterFunctionV2("generate_full_name", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
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
			res, err := generateRandomFullName(randomizer, maxLength)
			return res, err
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

/* Generates a random full name */
func generateRandomFullName(randomizer *rand.Rand, maxLength int64) (string, error) {
	maxLengthMinusSpace := maxLength - 1
	if maxLengthMinusSpace <= 0 {
		return "", fmt.Errorf("unable to generate full name including space with provided max length: %d", maxLength)
	}
	maxFirstNameIdx, maxLastNameIdx := transformer_utils.FindClosestPair(
		transformers_dataset.FirstNameIndices, transformers_dataset.LastNameIndices,
		maxLengthMinusSpace,
	)
	if maxFirstNameIdx == -1 || maxLastNameIdx == -1 {
		return "", fmt.Errorf("unable to generate a full name with the provided max length: %d", maxLength)
	}

	maxFirstNameLength := transformers_dataset.FirstNameIndices[maxFirstNameIdx]
	maxLastNameLength := transformers_dataset.LastNameIndices[maxLastNameIdx]
	firstname, err := generateRandomFirstName(randomizer, nil, maxFirstNameLength)
	if err != nil {
		return "", fmt.Errorf("unable to generate random first name with length: %d", maxFirstNameLength)
	}
	lastname, err := generateRandomLastName(randomizer, nil, maxLastNameLength)
	if err != nil {
		return "", fmt.Errorf("unable to generate random last name with length: %d", maxLastNameLength)
	}

	return fmt.Sprintf("%s %s", firstname, lastname), nil
}
