package transformers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/google/uuid"
)

// +neosyncTransformerBuilder:generate:generateSHA256Hash

func init() {
	spec := bloblang.NewPluginSpec()

	err := bloblang.RegisterFunctionV2("generate_sha256hash", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		return func() (any, error) {
			val, err := generateRandomSHA256Hash(uuid.NewString())
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

func (t *GenerateSHA256Hash) Generate(opts any) (any, error) {
	return generateRandomSHA256Hash(uuid.NewString())
}

func generateRandomSHA256Hash(input string) (string, error) {
	bites := []byte(input)
	hasher := sha256.New()
	_, err := hasher.Write(bites)
	if err != nil {
		return "", err
	}
	hash := hasher.Sum(nil)
	return hex.EncodeToString(hash), nil
}
