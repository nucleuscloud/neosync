package transformers

import (
	"fmt"
	"testing"
	"time"

	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/stretchr/testify/assert"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

func Test_GenerateStreetAddress(t *testing.T) {
	res, err := generateRandomStreetAddress(rng.New(time.Now().UnixMilli()), maxLength)
	assert.NoError(t, err)

	assert.IsType(t, "", res, "The returned street address should be a string")

	assert.LessOrEqual(t, int64(len(res)), maxLength, fmt.Sprintf("The city should be less than or equal to the max length. This is the error street address:%s", res))
}

func Test_GenerateStreetAddressShortMax(t *testing.T) {
	res, err := generateRandomStreetAddress(rng.New(time.Now().UnixMilli()), int64(5))
	assert.NoError(t, err)

	assert.IsType(t, "", res, "The returned street address should be a string")

	assert.LessOrEqual(t, int64(len(res)), maxLength, fmt.Sprintf("The city should be less than or equal to the max length. This is the error street address:%s", res))
}

func Test_GenerateStreetAddressSVeryhortMax(t *testing.T) {
	res, err := generateRandomStreetAddress(rng.New(time.Now().UnixMilli()), int64(2))
	assert.NoError(t, err)

	assert.IsType(t, "", res, "The returned street address should be a string")

	assert.LessOrEqual(t, int64(len(res)), maxLength, fmt.Sprintf("The city should be less than or equal to the max length. This is the error street address:%s", res))
}

func Test_StreetAddressTransformer(t *testing.T) {
	mapping := fmt.Sprintf(`root = generate_street_address(max_length:%d)`, maxLength)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the street address transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.IsType(t, Address{}.Address1, res, "The returned street address should be a string")
	assert.LessOrEqual(t, int64(len(res.(string))), maxLength, fmt.Sprintf("The city should be less than or equal to the max length. This is the error street address:%s", res))
}

func Test_StreetAddressTransformer_NoOptions(t *testing.T) {
	mapping := `root = generate_street_address()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the street address transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
}
