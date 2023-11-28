package transformers

import (
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformers_dataset "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/data-sets"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateCity(t *testing.T) {

	res := GenerateRandomCity()

	assert.IsType(t, "", res, "The returned city should be a string")

	cityExists := false
	for _, address := range transformers_dataset.Addresses {
		if address.City == res {
			cityExists = true
			break
		}
	}

	assert.True(t, cityExists, "The generated city should exist in the addresses array")
}

func Test_CityTransformer(t *testing.T) {
	mapping := `root = generate_city()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the city transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.IsType(t, "", res, "The returned city should be a string")

	cityExists := false
	for _, address := range transformers_dataset.Addresses {
		if address.City == res {
			cityExists = true
			break
		}
	}

	assert.True(t, cityExists, "The generated city should exist in the addresses array")
}
