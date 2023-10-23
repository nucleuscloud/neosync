package neosync_transformers

import (
	"testing"
	"time"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func TestProcessUTCTimestamp(t *testing.T) {

	timestamp, err := GenerateRandomUTCTimestamp()
	assert.NoError(t, err, "Error generating random UTC timestamp")

	// Check if the timestamp's time zone is UTC
	location := timestamp.Location()
	assert.Equal(t, location, time.UTC, "Generated timestamp is not in UTC")
}

func TestUTCTimestampTransformer(t *testing.T) {
	mapping := `root = utctimestamptransformer()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the random utc timestamp transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	location := res.(time.Time).Location()
	assert.Equal(t, location, time.UTC, "Generated timestamp is not in UTC")
}
