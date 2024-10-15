package transformers

import (
	"testing"
	"time"

	transformers_dataset "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/data-sets"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/stretchr/testify/assert"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

func Test_GenerateCountry(t *testing.T) {
	res := generateRandomCountry(rng.New(time.Now().UnixMilli()), false)

	assert.IsType(t, "", res, "The returned country should be a string")

	countryExists := false
	for _, address := range transformers_dataset.Countries {
		if address.Code == res {
			countryExists = true
			break
		}
	}

	assert.True(t, countryExists, "The generated country should exist in the country_codes.go file")
}

func Test_GenerateCountryCodeLength(t *testing.T) {
	res := generateRandomCountry(rng.New(time.Now().UnixMilli()), false)

	assert.IsType(t, "", res, "The returned country should be a string")

	countryExists := false
	for _, country := range transformers_dataset.Countries {
		if country.Code == res {
			countryExists = true
			break
		}
	}

	assert.Len(t, res, 2)
	assert.True(t, countryExists, "The generated country should exist in the countrys.go file")
}

func Test_GenerateCountryCodeFullName(t *testing.T) {
	res := generateRandomCountry(rng.New(time.Now().UnixMilli()), true)

	assert.IsType(t, "", res, "The returned country should be a string")

	countryExists := false
	for _, country := range transformers_dataset.Countries {
		if country.FullName == res {
			countryExists = true
			break
		}
	}

	assert.True(t, len(res) > 2)
	assert.True(t, countryExists, "The generated country should exist in the countrys.go file")
}

func Test_CountryTransformer(t *testing.T) {
	mapping := `root = generate_country(generate_full_name:false)`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the country transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.IsType(t, Address{}.City, res, "The returned country should be a string")

	countryExists := false
	for _, country := range transformers_dataset.Countries {
		if country.Code == res {
			countryExists = true
			break
		}
	}

	assert.Len(t, res, 2)
	assert.True(t, countryExists, "The generated country should exist in the countrys.go file")
}

func Test_CountryTransformer_NoOptions(t *testing.T) {
	mapping := `root = generate_country()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the country transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.IsType(t, Address{}.City, res, "The returned country should be a string")

	countryExists := false
	for _, country := range transformers_dataset.Countries {
		if country.Code == res {
			countryExists = true
			break
		}
	}

	assert.Len(t, res, 2)
	assert.True(t, countryExists, "The generated country should exist in the countrys.go file")
}
