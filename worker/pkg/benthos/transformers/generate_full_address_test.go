package transformers

import (
	"fmt"
	"testing"
	"time"

	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

var faMaxLength = int64(40)

func Test_GenerateFullAddress(t *testing.T) {
	res, err := generateRandomFullAddress(rng.New(time.Now().UnixMilli()), faMaxLength)
	assert.NoError(t, err)

	assert.IsType(t, "", res, "The returned full address should be a string")
	assert.LessOrEqual(t, int64(len(res)), faMaxLength, fmt.Sprintf("The city should be less than or equal to the max length. This is the error address:%s", res))
}

func Test_GenerateFullAddressVeryShortMax(t *testing.T) {
	shortMax := int64(15)

	res, err := generateRandomFullAddress(rng.New(time.Now().UnixMilli()), shortMax)
	assert.NoError(t, err)

	assert.IsType(t, "", res, "The returned full address should be a string")
	assert.LessOrEqual(t, int64(len(res)), shortMax, fmt.Sprintf("The city should be less than or equal to the max length. This is the error address:%s", res))
}

func Test_GenerateFullAddressShortMax(t *testing.T) {
	shortMax := int64(25)

	res, err := generateRandomFullAddress(rng.New(time.Now().UnixMilli()), shortMax)
	assert.NoError(t, err)

	assert.IsType(t, "", res, "The returned full address should be a string")
	assert.LessOrEqual(t, int64(len(res)), shortMax, fmt.Sprintf("The city should be less than or equal to the max length. This is the error address:%s", res))
}

func Test_FullAddressTransformer(t *testing.T) {
	mapping := fmt.Sprintf(`root = generate_full_address(max_length:%d)`, faMaxLength)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the full address transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.IsType(t, "", res, "The returned full address should be a string")
	assert.LessOrEqual(t, int64(len(res.(string))), faMaxLength, fmt.Sprintf("The city should be less than or equal to the max length. This is the error address:%s", res))
}

func Test_FullAddressTransformer_NoOptions(t *testing.T) {
	mapping := `root = generate_full_address()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the full address transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
}
