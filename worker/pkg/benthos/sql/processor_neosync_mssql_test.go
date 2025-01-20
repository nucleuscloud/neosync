package neosync_benthos_sql

import (
	"context"
	"testing"

	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/stretchr/testify/require"
	"github.com/warpstreamlabs/bento/public/service"
)

func Test_transformNeosyncToMssql(t *testing.T) {
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

		result, err := transformNeosyncToMssql(input, columns, columnDefaultProperties)
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

		result, err := transformNeosyncToMssql(input, columns, columnDefaultProperties)
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

		result, err := transformNeosyncToMssql(input, columns, columnDefaultProperties)
		require.NoError(t, err)

		require.Equal(t, 1, result["id"])
		require.Equal(t, "test", result["name"])
		_, exists := result["unknown_column"]
		require.False(t, exists)
	})

	t.Run("returns error for invalid root type", func(t *testing.T) {
		result, err := transformNeosyncToMssql("invalid", columns, columnDefaultProperties)
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

func Test_NeosyncToMssqlProcessor(t *testing.T) {
	conf := `
columns:
  - id
  - name
  - age
  - balance
  - is_active
  - created_at
  - default_value
column_data_types:
  id: integer
  name: text
  age: integer
  balance: double
  is_active: boolean
  created_at: timestamp
  default_value: text
column_default_properties:
  id:
    has_default_transformer: false
  name:
    has_default_transformer: false
  default_value:
    has_default_transformer: true
`
	spec := neosyncToMssqlProcessorConfig()
	env := service.NewEnvironment()

	procConfig, err := spec.ParseYAML(conf, env)
	require.NoError(t, err)

	proc, err := newNeosyncToMssqlProcessor(procConfig, service.MockResources())
	require.NoError(t, err)

	msgMap := map[string]any{
		"id":            1,
		"name":          "test",
		"age":           30,
		"balance":       1000.50,
		"is_active":     true,
		"created_at":    "2023-01-01T00:00:00Z",
		"default_value": "some default",
	}
	msg := service.NewMessage(nil)
	msg.SetStructured(msgMap)
	batch := service.MessageBatch{
		msg,
	}

	results, err := proc.ProcessBatch(context.Background(), batch)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Len(t, results[0], 1)

	val, err := results[0][0].AsStructured()
	require.NoError(t, err)

	expected := map[string]any{
		"id":         msgMap["id"],
		"name":       msgMap["name"],
		"age":        msgMap["age"],
		"balance":    msgMap["balance"],
		"is_active":  msgMap["is_active"],
		"created_at": msgMap["created_at"],
	}
	require.Equal(t, expected, val)

	require.NoError(t, proc.Close(context.Background()))
}

func Test_NeosyncToMssqlProcessor_SubsetColumns(t *testing.T) {
	conf := `
columns:
  - id
  - name
column_data_types:
  id: integer
  name: text
  age: integer
  balance: double
  is_active: boolean
  created_at: timestamp
column_default_properties:
  id:
    has_default_transformer: false
  name:
    has_default_transformer: false
`
	spec := neosyncToMssqlProcessorConfig()
	env := service.NewEnvironment()

	procConfig, err := spec.ParseYAML(conf, env)
	require.NoError(t, err)

	proc, err := newNeosyncToMssqlProcessor(procConfig, service.MockResources())
	require.NoError(t, err)

	msgMap := map[string]any{
		"id":         1,
		"name":       "test",
		"age":        30,
		"balance":    1000.50,
		"is_active":  true,
		"created_at": "2023-01-01T00:00:00Z",
	}
	msg := service.NewMessage(nil)
	msg.SetStructured(msgMap)
	batch := service.MessageBatch{
		msg,
	}

	results, err := proc.ProcessBatch(context.Background(), batch)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Len(t, results[0], 1)

	val, err := results[0][0].AsStructured()
	require.NoError(t, err)

	expected := map[string]any{
		"id":   msgMap["id"],
		"name": msgMap["name"],
	}
	require.Equal(t, expected, val)

	require.NoError(t, proc.Close(context.Background()))
}
