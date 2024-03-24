package transformers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/google/uuid"
)

func init() {
	spec := bloblang.NewPluginSpec()

	err := bloblang.RegisterFunctionV2("generate_sha256hash", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		return func() (any, error) {
			val, err := generateRandomSHA256Hash()
			if err != nil {
				return false, fmt.Errorf("unable to run generate_sha256hash: %w", err)
			}
			return val, nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

/* Generates a random SHA256 hashed value */
func generateRandomSHA256Hash() (string, error) {
	input := uuid.NewString()

	bites := []byte(input)
	hasher := sha256.New()
	_, err := hasher.Write(bites)
	if err != nil {
		return "", err
	}

	hash := hasher.Sum(nil)
	return hex.EncodeToString(hash), nil
}
