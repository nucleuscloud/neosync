package transformers

import (
	"fmt"
	"testing"
	"time"

	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

var maxLength = int64(20)

func Test_GenerateCity(t *testing.T) {
	res, err := generateRandomCity(rng.New(time.Now().UnixMilli()), maxLength)
	assert.NoError(t, err)

	assert.IsType(t, "", res, "The returned city should be a string")
	assert.LessOrEqual(t, int64(len(res)), maxLength, fmt.Sprintf("The city should be less than or equal to the max length. This is the error city:%s", res))
}

func Test_GenerateCityShortMax(t *testing.T) {
	res, err := generateRandomCity(rng.New(time.Now().UnixMilli()), int64(3))
	assert.NoError(t, err)

	assert.IsType(t, "", res, "The returned city should be a string")
	assert.LessOrEqual(t, int64(len(res)), int64(3), fmt.Sprintf("The city should be less than or equal to the max length. This is the error city:%s", res))
}

func Test_CityTransformer(t *testing.T) {
	mapping := fmt.Sprintf(`root = generate_city(max_length:%d)`, maxLength)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the city transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.IsType(t, "", res, "The returned city should be a string")

	assert.LessOrEqual(t, int64(len(res.(string))), maxLength, fmt.Sprintf("The city should be less than or equal to the max length. This is the error city:%s", res))
}

func Test_CityTransformer_NoOptions(t *testing.T) {
	mapping := `root = generate_city()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the city transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.IsType(t, "", res, "The returned city should be a string")

	assert.NotEmptyf(t, res, fmt.Sprintf("The city should be not be empty. This is the error city:%s", res))
}
