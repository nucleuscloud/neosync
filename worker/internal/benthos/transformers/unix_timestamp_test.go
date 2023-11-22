package transformers

import (
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func TestProcessUnixTimestamp(t *testing.T) {

	timestamp, err := GenerateRandomUnixTimestamp()
	assert.NoError(t, err, "Error generating random unix timestamp")

	assert.True(t, timestamp > 0, "Generated timestamp is not a valid Unix timestamp")
}

func TestUnixTimestampTransformer(t *testing.T) {
	mapping := `root = unixtimestamptransformer()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the random unix timestamp transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.True(t, res.(int64) > 0, "Generated timestamp is not a valid Unix timestamp")
}
