package transformers

import (
	"testing"
	"time"

	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/require"
)

func Test_ProcessUTCTimestamp(t *testing.T) {
	timestamp := generateRandomUTCTimestamp(rng.New(time.Now().UnixNano()))

	// Check if the timestamp's time zone is UTC
	location := timestamp.Location()
	require.Equal(t, location, time.UTC, "Generated timestamp is not in UTC")
}

func Test_UTCTimestampTransformer(t *testing.T) {
	mapping := `root = generate_utctimestamp()`
	ex, err := bloblang.Parse(mapping)
	require.NoError(t, err, "failed to parse the random utc timestamp transformer")

	res, err := ex.Query(nil)
	require.NoError(t, err)

	location := res.(time.Time).Location()
	require.Equal(t, location, time.UTC, "Generated timestamp is not in UTC")
}
