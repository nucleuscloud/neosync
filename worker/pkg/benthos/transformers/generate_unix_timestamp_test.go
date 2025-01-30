package transformers

import (
	"testing"
	"time"

	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/require"
)

func Test_GenerateUnixTimestamp(t *testing.T) {
	timestamp := generateRandomUnixTimestamp(rng.New(time.Now().UnixNano()))
	require.True(t, timestamp > 0, "Generated timestamp is not a valid Unix timestamp")
}

func Test_UnixTimestampTransformer(t *testing.T) {
	mapping := `root = generate_unixtimestamp()`
	ex, err := bloblang.Parse(mapping)
	require.NoError(t, err, "failed to parse the random unix timestamp transformer")

	res, err := ex.Query(nil)
	require.NoError(t, err)

	require.True(t, res.(int64) > 0, "Generated timestamp is not a valid Unix timestamp")
}
