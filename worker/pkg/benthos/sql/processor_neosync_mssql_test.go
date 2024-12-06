package neosync_benthos_sql

import (
	"testing"

	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/stretchr/testify/require"
	"github.com/warpstreamlabs/bento/public/service"
)

func Test_transformNeosyncToMssql(t *testing.T) {
	logger := &service.Logger{}
	columns := []string{"id", "name", "data", "default_col"}
	columnDefaultProperties := map[string]*neosync_benthos.ColumnDefaultProperties{
		"default_col": {HasDefaultTransformer: true},
	}

	t.Run("handles basic values", func(t *testing.T) {
		input := map[string]any{
			"id":          1,
			"name":        "test",
			"data":        map[string]string{"foo": "bar"},
			"default_col": "should be skipped",
		}

		result, err := transformNeosyncToMssql(logger, input, columns, columnDefaultProperties)
		require.NoError(t, err)

		require.Equal(t, 1, result["id"])
		require.Equal(t, "test", result["name"])
		require.Equal(t, []byte(`{"foo":"bar"}`), result["data"])
		_, exists := result["default_col"]
		require.False(t, exists)
	})

	t.Run("handles nil values", func(t *testing.T) {
		input := map[string]any{
			"id":   nil,
			"name": nil,
		}

		result, err := transformNeosyncToMssql(logger, input, columns, columnDefaultProperties)
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

		result, err := transformNeosyncToMssql(logger, input, columns, columnDefaultProperties)
		require.NoError(t, err)

		require.Equal(t, 1, result["id"])
		require.Equal(t, "test", result["name"])
		_, exists := result["unknown_column"]
		require.False(t, exists)
	})

	t.Run("returns error for invalid root type", func(t *testing.T) {
		result, err := transformNeosyncToMssql(logger, "invalid", columns, columnDefaultProperties)
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "root value must be a map[string]any")
	})
}

func Test_getMssqlValue(t *testing.T) {
	t.Run("marshals json for map value", func(t *testing.T) {
		input := map[string]string{"foo": "bar"}
		result, err := getMssqlValue(input)
		require.NoError(t, err)
		require.Equal(t, []byte(`{"foo":"bar"}`), result)
	})

	t.Run("returns original value for non-map types", func(t *testing.T) {
		result, err := getMssqlValue("test")
		require.NoError(t, err)
		require.Equal(t, "test", result)
	})

	t.Run("handles nil value", func(t *testing.T) {
		result, err := getMssqlValue(nil)
		require.NoError(t, err)
		require.Nil(t, result)
	})
}
