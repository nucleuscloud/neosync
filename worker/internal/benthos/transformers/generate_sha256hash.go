package transformers

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

func init() {

	spec := bloblang.NewPluginSpec()

	err := bloblang.RegisterFunctionV2("generate_sha256hash", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		return func() (any, error) {

			val, err := GenerateRandomSHA256Hash()

			if err != nil {
				return false, fmt.Errorf("unable to generate sha256 hash")
			}
			return val, nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

/* Generates a random SHA256 hashed value */
func GenerateRandomSHA256Hash() (string, error) {

	min := int64(1)
	max := int64(9)

	str, err := transformer_utils.GenerateRandomStringWithInclusiveBounds(min, max)
	if err != nil {
		return "", err
	}

	// hash the value
	bites := []byte(str)
	hasher := sha256.New()
	_, err = hasher.Write(bites)
	if err != nil {
		return "", err
	}

	// compute sha256 checksum and encode it into a hex string
	hashed := hasher.Sum(nil)
	var buf bytes.Buffer
	e := hex.NewEncoder(&buf)
	_, err = e.Write(hashed)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
