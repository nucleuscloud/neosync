package neosync_benthos_sql

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"testing"

	"github.com/doug-martin/goqu/v9"
	"github.com/lib/pq"
	neosynctypes "github.com/nucleuscloud/neosync/internal/neosync-types"
	pgutil "github.com/nucleuscloud/neosync/internal/postgres"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/stretchr/testify/require"
	"github.com/warpstreamlabs/bento/public/service"
)

func Test_getValidJson(t *testing.T) {
	t.Run("already valid json", func(t *testing.T) {
		input := []byte(`{"key": "value"}`)
		got, err := getValidJson(input)
		require.NoError(t, err)
		require.Equal(t, input, got)
	})

	t.Run("unquoted string", func(t *testing.T) {
		input := []byte(`hello world`)
		got, err := getValidJson(input)
		require.NoError(t, err)
		expected := []byte(`"hello world"`)
		require.Equal(t, expected, got)
	})
}

func Test_stringifyJsonArray(t *testing.T) {
	t.Run("array of objects", func(t *testing.T) {
		input := []any{
			map[string]any{"name": "Alice"},
			map[string]any{"name": "Bob"},
		}
		got, err := stringifyJsonArray(input)
		require.NoError(t, err)
		expected := []string{`{"name":"Alice"}`, `{"name":"Bob"}`}
		require.Equal(t, expected, got)
	})

	t.Run("empty array", func(t *testing.T) {
		got, err := stringifyJsonArray([]any{})
		require.NoError(t, err)
		require.Equal(t, []string{}, got)
	})
}

func Test_isColumnInList(t *testing.T) {
	columns := []string{"id", "name", "email"}

	t.Run("column exists", func(t *testing.T) {
		require.True(t, isColumnInList("name", columns))
	})

	t.Run("column does not exist", func(t *testing.T) {
		require.False(t, isColumnInList("age", columns))
	})

	t.Run("empty column list", func(t *testing.T) {
		require.False(t, isColumnInList("name", []string{}))
	})
}

func Test_processPgArrayFromJson(t *testing.T) {
	t.Run("json array", func(t *testing.T) {
		input := []byte(`[{"tag":"cool"},{"tag":"awesome"}]`)
		got, err := processPgArrayFromJson(input, "json[]")
		require.NoError(t, err)

		// Convert back to string for comparison since pq.Array isn't easily comparable
		arr, ok := got.(interface{ Value() (driver.Value, error) })
		require.True(t, ok)
		val, err := arr.Value()
		require.NoError(t, err)
		strArr, ok := val.(string)
		require.True(t, ok)
		require.Equal(t, `{"{\"tag\":\"cool\"}","{\"tag\":\"awesome\"}"}`, strArr)
	})

	t.Run("invalid json", func(t *testing.T) {
		input := []byte(`[invalid json]`)
		_, err := processPgArrayFromJson(input, "json[]")
		require.Error(t, err)
	})
}

func Test_transformNeosyncToPgx(t *testing.T) {
	logger := &service.Logger{}
	columns := []string{"id", "name", "data"}
	columnDataTypes := map[string]string{
		"id":   "integer",
		"name": "text",
		"data": "json",
	}
	columnDefaultProperties := map[string]*neosync_benthos.ColumnDefaultProperties{
		"id": {HasDefaultTransformer: true},
	}

	t.Run("transforms values correctly", func(t *testing.T) {
		input := map[string]any{
			"id":   123,
			"name": "test",
			"data": map[string]string{"foo": "bar"},
		}

		got, err := transformNeosyncToPgx(logger, input, columns, columnDataTypes, columnDefaultProperties)
		require.NoError(t, err)

		// id should be DEFAULT due to HasDefaultTransformer
		idVal, ok := got["id"].(goqu.Expression)
		require.True(t, ok)
		require.NotNil(t, idVal)

		require.Equal(t, "test", got["name"])

		// data should be JSON encoded
		dataBytes, ok := got["data"].([]byte)
		require.True(t, ok)
		require.JSONEq(t, `{"foo":"bar"}`, string(dataBytes))
	})

	t.Run("skips columns not in list", func(t *testing.T) {
		input := map[string]any{
			"id":      123,
			"name":    "test",
			"ignored": "value",
		}

		got, err := transformNeosyncToPgx(logger, input, columns, columnDataTypes, columnDefaultProperties)
		require.NoError(t, err)
		require.NotContains(t, got, "ignored")
	})

	t.Run("handles nil values", func(t *testing.T) {
		input := map[string]any{
			"id":   nil,
			"name": nil,
		}

		got, err := transformNeosyncToPgx(logger, input, columns, columnDataTypes, columnDefaultProperties)
		require.NoError(t, err)
		require.Nil(t, got["name"])
	})

	t.Run("invalid input type", func(t *testing.T) {
		_, err := transformNeosyncToPgx(logger, "not a map", columns, columnDataTypes, columnDefaultProperties)
		require.Error(t, err)
	})
}

func Test_getPgxValue(t *testing.T) {
	t.Run("handles json values", func(t *testing.T) {
		testCases := []struct {
			name     string
			input    any
			datatype string
			expected []byte
		}{
			{
				name:     "string value",
				input:    "value1",
				datatype: "json",
				expected: []byte(`"value1"`),
			},
			{
				name:     "number value",
				input:    42,
				datatype: "jsonb",
				expected: []byte(`42`),
			},
			{
				name:     "boolean value",
				input:    true,
				datatype: "json",
				expected: []byte(`true`),
			},
			{
				name:     "object value",
				input:    map[string]any{"key": "value"},
				datatype: "jsonb",
				expected: []byte(`{"key":"value"}`),
			},
			{
				name:     "array value",
				input:    []int{1, 2, 3},
				datatype: "json",
				expected: []byte(`[1,2,3]`),
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				got, err := getPgxValue(tc.input, nil, tc.datatype)
				require.NoError(t, err)
				require.Equal(t, tc.expected, got)
			})
		}
	})

	t.Run("handles default transformer", func(t *testing.T) {
		colDefaults := &neosync_benthos.ColumnDefaultProperties{
			HasDefaultTransformer: true,
		}
		got, err := getPgxValue("test", colDefaults, "text")
		require.NoError(t, err)
		require.Equal(t, goqu.Default(), got)
	})

	t.Run("handles nil value", func(t *testing.T) {
		got, err := getPgxValue(nil, nil, "text")
		require.NoError(t, err)
		require.Nil(t, got)
	})

	t.Run("handles byte slice", func(t *testing.T) {
		input := []byte("test")
		got, err := getPgxValue(input, nil, "text")
		require.NoError(t, err)
		require.Equal(t, input, got)
	})

	t.Run("handles slice", func(t *testing.T) {
		input := []string{"a", "b", "c"}
		got, err := getPgxValue(input, nil, "text[]")
		require.NoError(t, err)
		require.Equal(t, pq.Array(input), got)
	})

	t.Run("handles multidimensional slice", func(t *testing.T) {
		input := [][]string{{"a", "b"}, {"c", "d"}}
		got, err := getPgxValue(input, nil, "text[][]")
		require.NoError(t, err)
		require.Equal(t, goqu.Literal(pgutil.FormatPgArrayLiteral(input, "text[][]")), got)
	})

	t.Run("handles slice of maps", func(t *testing.T) {
		input := []map[string]string{{"key": "value"}}
		got, err := getPgxValue(input, nil, "jsonb[]")
		require.NoError(t, err)
		require.Equal(t, goqu.Literal(pgutil.FormatPgArrayLiteral(input, "jsonb[]")), got)
	})
}

func Test_NeosyncToPgxProcessor(t *testing.T) {
	conf := `
columns:
  - id
  - name
  - age
  - balance
  - is_active
  - created_at
  - tags
  - metadata
  - interval
  - default_value
column_data_types:
  id: integer
  name: text
  age: integer
  balance: double
  is_active: boolean
  created_at: timestamp
  tags: text[]
  metadata: jsonb
  interval: interval
  default_value: text
column_default_properties:
  id:
    has_default_transformer: false
  name:
    has_default_transformer: false
  default_value:
    has_default_transformer: true
`
	spec := neosyncToPgxProcessorConfig()
	env := service.NewEnvironment()

	procConfig, err := spec.ParseYAML(conf, env)
	require.NoError(t, err)

	proc, err := newNeosyncToPgxProcessor(procConfig, service.MockResources())
	require.NoError(t, err)

	interval, err := neosynctypes.NewInterval()
	require.NoError(t, err)
	interval.ScanPgx(map[string]any{
		"months":       1,
		"days":         10,
		"microseconds": 3600000000,
	})

	msgMap := map[string]any{
		"id":            1,
		"name":          "test",
		"age":           30,
		"balance":       1000.50,
		"is_active":     true,
		"created_at":    "2023-01-01T00:00:00Z",
		"tags":          []string{"tag1", "tag2"},
		"metadata":      map[string]string{"key": "value"},
		"interval":      interval,
		"default_value": "some value",
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

	intervalVal, err := interval.ValuePgx()
	jsonBytes, err := json.Marshal(msgMap["metadata"])
	require.NoError(t, err)

	require.NoError(t, err)
	expected := map[string]any{
		"id":            msgMap["id"],
		"name":          msgMap["name"],
		"age":           msgMap["age"],
		"balance":       msgMap["balance"],
		"is_active":     msgMap["is_active"],
		"created_at":    msgMap["created_at"],
		"tags":          pq.Array(msgMap["tags"]),
		"metadata":      jsonBytes,
		"interval":      intervalVal,
		"default_value": goqu.Default(),
	}
	require.Equal(t, expected, val)

	require.NoError(t, proc.Close(context.Background()))
}

func Test_NeosyncToPgxProcessor_SubsetColumns(t *testing.T) {
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
  tags: text[]
  metadata: jsonb
  interval: interval
column_default_properties:
  id:
    has_default_transformer: false
  name:
    has_default_transformer: false
`
	spec := neosyncToPgxProcessorConfig()
	env := service.NewEnvironment()

	procConfig, err := spec.ParseYAML(conf, env)
	require.NoError(t, err)

	proc, err := newNeosyncToPgxProcessor(procConfig, service.MockResources())
	require.NoError(t, err)

	msgMap := map[string]any{
		"id":         1,
		"name":       "test",
		"age":        30,
		"balance":    1000.50,
		"is_active":  true,
		"created_at": "2023-01-01T00:00:00Z",
		"tags":       []string{"tag1", "tag2"},
		"metadata":   map[string]string{"key": "value"},
		"interval": neosynctypes.Interval{
			Months:       1,
			Days:         10,
			Microseconds: 3600000000,
		},
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
		"id":   1,
		"name": "test",
	}
	require.Equal(t, expected, val)

	require.NoError(t, proc.Close(context.Background()))
}
