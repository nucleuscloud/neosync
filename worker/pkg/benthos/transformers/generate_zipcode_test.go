package transformers

import (
	"testing"
	"time"

	transformers_dataset "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/data-sets"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/stretchr/testify/assert"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

func Test_GenerateZipcode(t *testing.T) {
	res := generateRandomZipcode(rng.New(time.Now().UnixMilli()))

	assert.IsType(t, "", res, "The returned zipcode should be a string")

	zipcodeExists := false
	for _, address := range transformers_dataset.Addresses {
		if address.Zipcode == res {
			zipcodeExists = true
			break
		}
	}

	assert.True(t, zipcodeExists, "The generated zipcode should exist in the addresses array")
}

func Test_ZipcodeTransformer(t *testing.T) {
	mapping := `root = generate_zipcode()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the zipcode transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.IsType(t, Address{}.City, res, "The returned zipcode should be a string")

	zipcodeExists := false
	for _, address := range transformers_dataset.Addresses {
		if address.Zipcode == res {
			zipcodeExists = true
			break
		}
	}

	assert.True(t, zipcodeExists, "The generated zipcode should exist in the addresses array")
}
