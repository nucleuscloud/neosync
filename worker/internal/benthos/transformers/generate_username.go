package transformers

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/benthosdev/benthos/v4/public/bloblang"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

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
	// randomly select a letter in the alphabet to use as a first initial
	fn := string(alphabet[randomizer.Intn(len(alphabet))])

	ln, err := generateRandomLastName(randomizer, nil, maxLength-1)
	if err != nil {
		return "", err
	}
	return fn + strings.ToLower(ln), nil
}
