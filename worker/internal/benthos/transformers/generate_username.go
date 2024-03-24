package transformers

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

func init() {
	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewInt64Param("max_length").Default(10000)).
		Param(bloblang.NewInt64Param("seed").Default(time.Now().UnixNano()))

	err := bloblang.RegisterFunctionV2("generate_username", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		return func() (any, error) {
			maxLength, err := args.GetInt64("max_length")
			if err != nil {
				return nil, err
			}
			seed, err := args.GetInt64("seed")
			if err != nil {
				return nil, err
			}
			randomizer := rand.New(rand.NewSource(seed)) //nolint:gosec

			res, err := generateUsername(randomizer, maxLength)
			if err != nil {
				return nil, fmt.Errorf("unable to run generate_username: %w", err)
			}
			return res, nil
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

// Generates a username with a lowercase first initial and titlecase lastname
func generateUsername(randomizer *rand.Rand, maxLength int64) (string, error) {
	fn := transformer_utils.GetRandomCharacterString(randomizer, 1)

	ln, err := generateRandomLastName(randomizer, nil, maxLength-1)
	if err != nil {
		return "", err
	}
	return fn + strings.ToLower(ln), nil
}
