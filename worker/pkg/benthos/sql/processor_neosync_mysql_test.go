package neosync_benthos_sql

import (
	"testing"

	"github.com/doug-martin/goqu/v9"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/stretchr/testify/require"
	"github.com/warpstreamlabs/bento/public/service"
)

func Test_transformNeosyncToMysql(t *testing.T) {
	logger := &service.Logger{}
	columns := []string{"id", "name", "data", "bits", "default_col"}
	columnDataTypes := map[string]string{
		"id":   "int",
		"name": "varchar",
		"data": "json",
		"bits": "bit",
	}
	columnDefaultProperties := map[string]*neosync_benthos.ColumnDefaultProperties{
		"default_col": {HasDefaultTransformer: true},
	}

	t.Run("handles basic values", func(t *testing.T) {
		input := map[string]any{
			"id":          1,
			"name":        "test",
			"data":        map[string]string{"foo": "bar"},
			"bits":        []byte("1"),
			"default_col": "should be default",
		}

		result, err := transformNeosyncToMysql(logger, input, columns, columnDataTypes, columnDefaultProperties)
		require.NoError(t, err)

		require.Equal(t, 1, result["id"])
		require.Equal(t, "test", result["name"])
		require.Equal(t, []byte(`{"foo":"bar"}`), result["data"])
		require.Equal(t, []byte{1}, result["bits"])
		require.Equal(t, goqu.Default(), result["default_col"])
	})

	t.Run("handles nil values", func(t *testing.T) {
		input := map[string]any{
			"id":   nil,
			"name": nil,
		}

		result, err := transformNeosyncToMysql(logger, input, columns, columnDataTypes, columnDefaultProperties)
		require.NoError(t, err)

		require.Nil(t, result["id"])
		require.Nil(t, result["name"])
	})

	t.Run("skips columns not in column list", func(t *testing.T) {
		input := map[string]any{
			"id":             1,
			"name":           "test",
			"unknown_column": "should not appear",
		}

		result, err := transformNeosyncToMysql(logger, input, columns, columnDataTypes, columnDefaultProperties)
		require.NoError(t, err)

		require.Equal(t, 1, result["id"])
		require.Equal(t, "test", result["name"])
		_, exists := result["unknown_column"]
		require.False(t, exists)
	})

	t.Run("returns error for invalid root type", func(t *testing.T) {
		result, err := transformNeosyncToMysql(logger, "invalid", columns, columnDataTypes, columnDefaultProperties)
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "root value must be a map[string]any")
	})
}

func Test_getMysqlValue(t *testing.T) {
	t.Run("returns default for column with default transformer", func(t *testing.T) {
		colDefaults := &neosync_benthos.ColumnDefaultProperties{HasDefaultTransformer: true}
		result, err := getMysqlValue("test", colDefaults, "varchar")
		require.NoError(t, err)
		require.Equal(t, goqu.Default(), result)
	})

	t.Run("marshals json for json datatype", func(t *testing.T) {
		input := map[string]string{"foo": "bar"}
		result, err := getMysqlValue(input, nil, "json")
		require.NoError(t, err)
		require.Equal(t, []byte(`{"foo":"bar"}`), result)
	})

	t.Run("handles bit datatype", func(t *testing.T) {
		result, err := getMysqlValue([]byte("1"), nil, "bit")
		require.NoError(t, err)
		require.Equal(t, []byte{1}, result)
	})

	t.Run("returns original value for non-special cases", func(t *testing.T) {
		result, err := getMysqlValue("test", nil, "varchar")
		require.NoError(t, err)
		require.Equal(t, "test", result)
	})
}

func Test_handleMysqlByteSlice(t *testing.T) {
	t.Run("converts bit string to bytes", func(t *testing.T) {
		result, err := handleMysqlByteSlice([]byte("1"), "bit")
		require.NoError(t, err)
		require.Equal(t, []byte{1}, result)
	})

	t.Run("returns original bytes for non-bit type", func(t *testing.T) {
		input := []byte("test")
		result, err := handleMysqlByteSlice(input, "varchar")
		require.NoError(t, err)
		require.Equal(t, input, result)
	})
}
