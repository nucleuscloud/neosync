package mysql

import (
	"testing"

	neosynctypes "github.com/nucleuscloud/neosync/internal/neosync-types"
	"github.com/stretchr/testify/require"
)

func Test_parseMysqlRowValues(t *testing.T) {
	t.Run("Multiple Columns", func(t *testing.T) {
		binaryData := []byte{0x01, 0x02, 0x03}

		values := []any{
			"Hello",
			int64(42),
			true,
			nil,
			[]byte(`{"key": "value"}`),
			binaryData,
		}
		columnNames := []string{
			"text_col", "int_col", "bool_col", "nil_col", "json_col", "binary_col",
			"binary_col",
		}
		cTypes := []string{
			"text",
			"integer",
			"boolean",
			"text",
			"json",
			"binary",
		}
		result := parseMysqlRowValues(values, columnNames, cTypes)
		expected := map[string]any{
			"text_col": "Hello",
			"int_col":  int64(42),
			"bool_col": true,
			"nil_col":  nil,
			"json_col": map[string]any{"key": "value"},
			"binary_col": &neosynctypes.Binary{
				BaseType: neosynctypes.BaseType{
					Neosync: neosynctypes.Neosync{
						Version: 1,
						TypeId:  "NEOSYNC_BINARY",
					},
				},
				Bytes: binaryData,
			},
		}
		require.Equal(t, expected, result)
	})

	t.Run("JSON Columns", func(t *testing.T) {
		values := []any{[]byte(`"Hello"`), []byte(`true`), []byte(`null`), []byte(`42`), []byte(`{"items": ["book", "pen"], "count": 2, "in_stock": true}`), []byte(`[1,2,3]`)}
		columnNames := []string{"text_col", "bool_col", "null_col", "int_col", "json_col", "array_col"}
		cTypes := []string{"json", "json", "json", "json", "json", "json"}

		result := parseMysqlRowValues(values, columnNames, cTypes)

		expected := map[string]any{
			"text_col":  "Hello",
			"bool_col":  true,
			"null_col":  nil,
			"int_col":   float64(42),
			"json_col":  map[string]any{"items": []any{"book", "pen"}, "count": float64(2), "in_stock": true},
			"array_col": []any{float64(1), float64(2), float64(3)},
		}
		require.Equal(t, expected, result)
	})
}
