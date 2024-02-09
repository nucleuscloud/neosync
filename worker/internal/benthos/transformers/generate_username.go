package transformers

import (
	"math/rand"
	"strings"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
)

var alphabet = "abcdefghijklmnopqrstuvwxyz"

func init() {
	spec := bloblang.NewPluginSpec().Param(bloblang.NewInt64Param("max_length"))

	err := bloblang.RegisterFunctionV2("generate_username", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		return func() (any, error) {
			maxLength, err := args.GetInt64("max_length")
			if err != nil {
				return nil, err
			}

			res, err := GenerateUsername(maxLength)

			return res, err
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

// Generates a username with a lowercase first initial and titlecase lastname
func GenerateUsername(maxLength int64) (string, error) {
	//nolint
	// randomly select a letter in the alphabet to use as a first initial
	fn := string(alphabet[rand.Intn(len(alphabet))])

	ln, err := GenerateRandomLastName(maxLength - 1)
	if err != nil {
		return "", err
	}
	return fn + strings.ToLower(ln), nil
}
