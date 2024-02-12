package transformers

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

var randomizer = rand.New(rand.NewSource(int64(20)))

func Test_GenerateStreetAddress(t *testing.T) {
	res, err := GenerateRandomStreetAddress(maxLength, randomizer)
	assert.NoError(t, err)

	assert.IsType(t, "", res, "The returned street address should be a string")

	assert.LessOrEqual(t, int64(len(res)), maxLength, fmt.Sprintf("The street should be less than or equal to the max length. This is the error street address:%s", res))
}

func Test_GenerateStreetAddressShortMax(t *testing.T) {
	res, err := GenerateRandomStreetAddress(int64(5), randomizer)
	assert.NoError(t, err)

	assert.IsType(t, "", res, "The returned street address should be a string")

	assert.LessOrEqual(t, int64(len(res)), maxLength, fmt.Sprintf("The street should be less than or equal to the max length. This is the error street address:%s", res))
}

func Test_GenerateStreetAddressSVeryhortMax(t *testing.T) {
	res, err := GenerateRandomStreetAddress(int64(2), randomizer)
	assert.NoError(t, err)

	assert.IsType(t, "", res, "The returned street address should be a string")

	assert.LessOrEqual(t, int64(len(res)), maxLength, fmt.Sprintf("The street should be less than or equal to the max length. This is the error street address:%s", res))
}

func Test_StreetAddressTransformer(t *testing.T) {
	mapping := fmt.Sprintf(`root = generate_street_address(max_length:%d: seed:%d)`, maxLength, randomizer)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the street address transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.IsType(t, Address{}.Address1, res, "The returned street address should be a string")
	assert.LessOrEqual(t, int64(len(res.(string))), maxLength, fmt.Sprintf("The street should be less than or equal to the max length. This is the error street address:%s", res))
}

func Test_StreetAddressTransformerNoSeed(t *testing.T) {

	var noSeed *int64
	mapping := fmt.Sprintf(`root = generate_street_address(max_length:%d: seed:%d)`, maxLength, noSeed)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the street address transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.IsType(t, Address{}.Address1, res, "The returned street address should be a string")
	assert.LessOrEqual(t, int64(len(res.(string))), maxLength, fmt.Sprintf("The street should be less than or equal to the max length. This is the error street address:%s", res))
}
