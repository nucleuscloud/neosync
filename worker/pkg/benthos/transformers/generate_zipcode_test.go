package transformers

import (
	"testing"
	"time"

	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/stretchr/testify/assert"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

func Test_GenerateZipcode(t *testing.T) {
	res, err := generateRandomZipcode(rng.New(time.Now().UnixMilli()))
	assert.NoError(t, err, "failed to generate zipcode")
	assert.IsType(t, "", res, "The returned zipcode should be a string")
}

func Test_ZipcodeTransformer(t *testing.T) {
	mapping := `root = generate_zipcode()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the zipcode transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)
	assert.IsType(t, "", res, "The returned zipcode should be a string")
}
