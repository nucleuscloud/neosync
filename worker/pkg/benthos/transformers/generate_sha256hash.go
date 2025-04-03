package transformers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
)

// +neosyncTransformerBuilder:generate:generateSHA256Hash

func init() {
	spec := bloblang.NewPluginSpec().
		Description("Generates a random SHA256 hash and returns it as a string.").
		Category("string")

	err := bloblang.RegisterFunctionV2(
		"generate_sha256hash",
		spec,
		func(args *bloblang.ParsedParams) (bloblang.Function, error) {
			return func() (any, error) {
				val, err := generateRandomSHA256Hash(uuid.NewString())
				if err != nil {
					return false, fmt.Errorf("unable to run generate_sha256hash: %w", err)
				}
				return val, nil
			}, nil
		},
	)
	if err != nil {
		panic(err)
	}
}

func NewGenerateSHA256HashOptsFromConfig(
	config *mgmtv1alpha1.GenerateSha256Hash,
) (*GenerateSHA256HashOpts, error) {
	return NewGenerateSHA256HashOpts()
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
