package transformers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

func Test_ProcessUTCTimestamp(t *testing.T) {
	timestamp, err := generateRandomUTCTimestamp()
	assert.NoError(t, err, "Error generating random UTC timestamp")

	// Check if the timestamp's time zone is UTC
	location := timestamp.Location()
	assert.Equal(t, location, time.UTC, "Generated timestamp is not in UTC")
}

func Test_UTCTimestampTransformer(t *testing.T) {
	mapping := `root = generate_utctimestamp()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the random utc timestamp transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	location := res.(time.Time).Location()
	assert.Equal(t, location, time.UTC, "Generated timestamp is not in UTC")
}
