package shared

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_splitTableKey(t *testing.T) {
	schema, table := SplitTableKey("foo")
	require.Equal(t, schema, "public")
	require.Equal(t, table, "foo")

	schema, table = SplitTableKey("neosync.foo")
	require.Equal(t, schema, "neosync")
	require.Equal(t, table, "foo")
}
