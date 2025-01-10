package transformers

import (
	"testing"
	"time"

	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/stretchr/testify/assert"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

func Test_GenerateCountry(t *testing.T) {
	res, err := generateRandomCountry(rng.New(time.Now().UnixMilli()), false)
	assert.NoError(t, err, "failed to generate country")

	assert.IsType(t, "", res, "The returned country should be a string")
}

func Test_GenerateCountryCodeLength(t *testing.T) {
	res, err := generateRandomCountry(rng.New(time.Now().UnixMilli()), false)
	assert.NoError(t, err, "failed to generate country")
	assert.IsType(t, "", res, "The returned country should be a string")
	assert.Len(t, res, 2)
}

func Test_GenerateCountryCodeFullName(t *testing.T) {
	res, err := generateRandomCountry(rng.New(time.Now().UnixMilli()), true)
	assert.NoError(t, err, "failed to generate country")
	assert.IsType(t, "", res, "The returned country should be a string")
	assert.True(t, len(res) > 2)
}

func Test_CountryTransformer(t *testing.T) {
	mapping := `root = generate_country(generate_full_name:false)`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the country transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.Len(t, res, 2)
}

func Test_CountryTransformer_NoOptions(t *testing.T) {
	mapping := `root = generate_country()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the country transformer")
	res, err := ex.Query(nil)
	assert.NoError(t, err)
	assert.Len(t, res, 2)
}
