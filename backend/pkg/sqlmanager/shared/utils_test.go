package sqlmanager_shared

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_getUniqueSchemaColMappings(t *testing.T) {
	mappings := GetUniqueSchemaColMappings(
		[]*DatabaseSchemaRow{
			{TableSchema: "public", TableName: "users", ColumnName: "id"},
			{TableSchema: "public", TableName: "users", ColumnName: "created_by"},
			{TableSchema: "public", TableName: "users", ColumnName: "updated_by"},

			{TableSchema: "neosync_api", TableName: "accounts", ColumnName: "id"},
		},
	)
	require.Contains(t, mappings, "public.users", "job mappings are a subset of the present database schemas")
	require.Contains(t, mappings, "neosync_api.accounts", "job mappings are a subset of the present database schemas")
	require.Contains(t, mappings["public.users"], "id", "")
	require.Contains(t, mappings["public.users"], "created_by", "")
	require.Contains(t, mappings["public.users"], "updated_by", "")
	require.Contains(t, mappings["neosync_api.accounts"], "id", "")
}

func Test_splitTableKey(t *testing.T) {
	schema, table := SplitTableKey("foo")
	require.Equal(t, schema, "public")
	require.Equal(t, table, "foo")

	schema, table = SplitTableKey("neosync.foo")
	require.Equal(t, schema, "neosync")
	require.Equal(t, table, "foo")
}

func Test_DedupeSlice(t *testing.T) {
	input := []string{"foo", "bar", "foo", "bar", "baz"}
	actual := DedupeSlice(input)
	require.Equal(t, []string{"foo", "bar", "baz"}, actual)
}
