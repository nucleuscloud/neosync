package neosync_transformers

import (
	"testing"
	"time"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func TestProcessUTCTimestamp(t *testing.T) {

	res, err := GenerateRandomUTCTimestamp()
	assert.NoError(t, err, "Error parsing generated timestamp")

	parsedTime, parseErr := time.Parse("2006-01-02 15:04:05 +0000", res)
	assert.NoError(t, parseErr, "Error parsing generated timestamp")

	// Check if the time zone of the parsed time is UTC
	_, offset := parsedTime.Zone()
	assert.Equal(t, offset, 0, "Generated timestamp is not in UTC")
}

func TestUTCTimestampTransformer(t *testing.T) {
	mapping := `root = utctimestamptransformer()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the random utc timestamp transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	parsedTime, parseErr := time.Parse("2006-01-02 15:04:05 +0000", res.(string))
	assert.NoError(t, parseErr, "Error parsing generated timestamp")

	// Check if the time zone of the parsed time is UTC
	_, offset := parsedTime.Zone()
	assert.Equal(t, offset, 0, "Generated timestamp is not in UTC")
}
