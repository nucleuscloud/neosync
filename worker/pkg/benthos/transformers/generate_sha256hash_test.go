package transformers

import (
	"testing"

	"github.com/redpanda-data/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/require"
)

func Test_GenerateSHA256Hash(t *testing.T) {
	res, err := generateRandomSHA256Hash("123")
	require.NoError(t, err)
	require.NotEmpty(t, res)
}

func Test_GenerateSHA256HashTransformer(t *testing.T) {
	mapping := `root = generate_sha256hash()`
	ex, err := bloblang.Parse(mapping)
	require.NoError(t, err, "failed to parse the random float transformer")

	res, err := ex.Query(nil)
	require.NoError(t, err)
	require.IsType(t, "", res, "The actual value should be a string")
}
