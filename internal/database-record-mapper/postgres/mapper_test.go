package postgres

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	neosynctypes "github.com/nucleuscloud/neosync/internal/neosync-types"
	"github.com/stretchr/testify/require"
)

func Test_parsePgRowValues(t *testing.T) {
	t.Run("Multiple Columns", func(t *testing.T) {
		binaryData := []byte{0x01, 0x02, 0x03}
		xmlStr := "<root><element>value</element></root>"
		uuidValue := "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"

		xmlVal := &sql.NullString{String: xmlStr, Valid: true}
		jsonVal := &NullableJSON{RawMessage: json.RawMessage(`{"key": "value"}`), Valid: true}

		values := []any{
			"Hello",
			int64(42),
			true,
			nil,
			jsonVal,
			&PgxArray[any]{
				Array:       pgtype.Array[any]{Elements: []any{1, 2, 3}},
				colDataType: "_integer",
			},
			binaryData,
			xmlVal,
			uuidValue,
		}
		columnNames := []string{
			"text_col", "int_col", "bool_col", "nil_col", "json_col", "array_col",
			"binary_col", "xml_col", "uuid_col",
		}
		columnTypes := []string{
			"text", "integer", "boolean", "text", "json", "_integer", "bytea", "xml", "uuid",
		}

		result, err := parsePgRowValues(values, columnNames, columnTypes)
		require.NoError(t, err)
		expected := map[string]any{
			"text_col":  "Hello",
			"int_col":   int64(42),
			"bool_col":  true,
			"nil_col":   nil,
			"json_col":  map[string]any{"key": "value"},
			"array_col": []any{1, 2, 3},
			"binary_col": &neosynctypes.Binary{
				BaseType: neosynctypes.BaseType{
					Neosync: neosynctypes.Neosync{
						Version: 1,
						TypeId:  "NEOSYNC_BINARY",
					},
				},
				Bytes: binaryData,
			},
			"xml_col":  xmlStr,
			"uuid_col": uuidValue,
		}
		require.Equal(t, expected, result)
	})

	t.Run("Bit and Binary Columns", func(t *testing.T) {
		// Test bit type
		bits := &pgtype.Bits{Bytes: []byte{0x05}, Len: 3, Valid: true} // binary 101

		// Test varbit type
		varbits := &pgtype.Bits{Bytes: []byte{0x0D}, Len: 4, Valid: true} // binary 1101

		// Test bytea
		byteaValue := []byte{0xDE, 0xAD, 0xBE, 0xEF}

		// Test array types
		bitArray := &PgxArray[*pgtype.Bits]{
			Array: pgtype.Array[*pgtype.Bits]{
				Elements: []*pgtype.Bits{
					{Bytes: []byte{0x05}, Len: 3, Valid: true}, // 101
					{Bytes: []byte{0x03}, Len: 2, Valid: true}, // 11
				},
			},
			colDataType: "_bit",
		}

		byteaArray := &PgxArray[[]byte]{
			Array: pgtype.Array[[]byte]{
				Elements: [][]byte{
					{0x01, 0x02},
					{0x03, 0x04},
				},
			},
			colDataType: "_bytea",
		}

		values := []any{
			bits,
			varbits,
			byteaValue,
			bitArray,
			byteaArray,
		}

		columnNames := []string{
			"bit_col",
			"varbit_col",
			"bytea_col",
			"bit_array_col",
			"bytea_array_col",
		}

		columnTypes := []string{
			"bit", "varbit", "bytea", "_bit", "_bytea",
		}

		result, err := parsePgRowValues(values, columnNames, columnTypes)
		require.NoError(t, err)

		expected := map[string]any{
			"bit_col": &neosynctypes.Bits{
				BaseType: neosynctypes.BaseType{
					Neosync: neosynctypes.Neosync{
						Version: 1,
						TypeId:  "NEOSYNC_BIT",
					},
				},
				Bytes: bits.Bytes,
				Len:   bits.Len,
			},
			"varbit_col": &neosynctypes.Bits{
				BaseType: neosynctypes.BaseType{
					Neosync: neosynctypes.Neosync{
						Version: 1,
						TypeId:  "NEOSYNC_BIT",
					},
				},
				Bytes: varbits.Bytes,
				Len:   varbits.Len,
			},
			"bytea_col": &neosynctypes.Binary{
				BaseType: neosynctypes.BaseType{
					Neosync: neosynctypes.Neosync{
						Version: 1,
						TypeId:  "NEOSYNC_BINARY",
					},
				},
				Bytes: byteaValue,
			},
			"bit_array_col": &neosynctypes.NeosyncArray{
				BaseType: neosynctypes.BaseType{
					Neosync: neosynctypes.Neosync{
						Version: 1,
						TypeId:  "NEOSYNC_ARRAY",
					},
				},
				Elements: []neosynctypes.NeosyncAdapter{
					&neosynctypes.Bits{
						BaseType: neosynctypes.BaseType{
							Neosync: neosynctypes.Neosync{
								Version: 1,
								TypeId:  "NEOSYNC_BIT",
							},
						},
						Bytes: []byte{0x05},
						Len:   3,
					},
					&neosynctypes.Bits{
						BaseType: neosynctypes.BaseType{
							Neosync: neosynctypes.Neosync{
								Version: 1,
								TypeId:  "NEOSYNC_BIT",
							},
						},
						Bytes: []byte{0x03},
						Len:   2,
					},
				},
			},
			"bytea_array_col": &neosynctypes.NeosyncArray{
				BaseType: neosynctypes.BaseType{
					Neosync: neosynctypes.Neosync{
						Version: 1,
						TypeId:  "NEOSYNC_ARRAY",
					},
				},
				Elements: []neosynctypes.NeosyncAdapter{
					&neosynctypes.Binary{
						BaseType: neosynctypes.BaseType{
							Neosync: neosynctypes.Neosync{
								Version: 1,
								TypeId:  "NEOSYNC_BINARY",
							},
						},
						Bytes: []byte{0x01, 0x02},
					},
					&neosynctypes.Binary{
						BaseType: neosynctypes.BaseType{
							Neosync: neosynctypes.Neosync{
								Version: 1,
								TypeId:  "NEOSYNC_BINARY",
							},
						},
						Bytes: []byte{0x03, 0x04},
					},
				},
			},
		}

		require.Equal(t, expected, result)
	})

	t.Run("JSON Columns", func(t *testing.T) {
		values := []any{
			&NullableJSON{RawMessage: json.RawMessage(`"Hello"`), Valid: true},
			&NullableJSON{RawMessage: json.RawMessage(`true`), Valid: true},
			&NullableJSON{Valid: false},
			&NullableJSON{RawMessage: json.RawMessage(`42`), Valid: true},
			&NullableJSON{RawMessage: json.RawMessage(`{"items": ["book", "pen"], "count": 2, "in_stock": true}`), Valid: true},
			&NullableJSON{RawMessage: json.RawMessage(`[1,2,3]`), Valid: true},
			nil,
			&NullableJSON{RawMessage: json.RawMessage(`null`), Valid: true},
		}
		columnNames := []string{"text_col", "bool_col", "null_col", "int_col", "json_col", "array_col", "nil_col", "null_json"}
		columnTypes := []string{"json", "json", "json", "json", "json", "_json", "json", "json"}

		result, err := parsePgRowValues(values, columnNames, columnTypes)
		require.NoError(t, err)

		expected := map[string]any{
			"text_col":  "Hello",
			"bool_col":  true,
			"null_col":  nil,
			"int_col":   float64(42),
			"json_col":  map[string]any{"items": []any{"book", "pen"}, "count": float64(2), "in_stock": true},
			"array_col": []any{float64(1), float64(2), float64(3)},
			"nil_col":   nil,
			"null_json": "null",
		}
		require.Equal(t, expected, result)
	})

	t.Run("Multiple Array Columns", func(t *testing.T) {
		binaryData1 := []byte{0x01, 0x02, 0x03}
		binaryData2 := []byte{0x04, 0x05, 0x06}
		xmlData1 := "<root><element>value1</element></root>"
		xmlData2 := "<root><element>value2</element></root>"
		uuidValue1 := "160075f6-4d6e-4040-b287-bd43677464fa"
		uuidValue2 := "5f4a4b96-a74e-4502-b05b-1d96fba90657"

		values := []any{
			&PgxArray[any]{
				Array:       pgtype.Array[any]{Elements: []any{"Hello", "World"}},
				colDataType: "_text",
			},
			&PgxArray[any]{
				Array:       pgtype.Array[any]{Elements: []any{int64(42), int64(43)}},
				colDataType: "_integer",
			},
			&PgxArray[any]{
				Array:       pgtype.Array[any]{Elements: []any{true, false}},
				colDataType: "_boolean",
			},
			&PgxArray[any]{
				Array: pgtype.Array[any]{Elements: []any{
					map[string]any{"key": "value1"}, map[string]any{"key": "value2"},
				}},
				colDataType: "_json",
			},
			&PgxArray[[]byte]{
				Array:       pgtype.Array[[]byte]{Elements: [][]byte{binaryData1, binaryData2}},
				colDataType: "_bytea",
			},
			&PgxArray[any]{
				Array: pgtype.Array[any]{Elements: []any{
					xmlData1,
					xmlData2,
				}},
				colDataType: "_xml",
			},
			&PgxArray[any]{
				Array:       pgtype.Array[any]{Elements: []any{uuidValue1, uuidValue2}},
				colDataType: "_uuid",
			},
			&PgxArray[any]{
				Array:       pgtype.Array[any]{Elements: []any{[]any{1, 2}, []any{3, 4}}},
				colDataType: "_integer[]",
			},
		}

		columnNames := []string{
			"text_array", "int_array", "bool_array", "json_array",
			"binary_array", "xml_array", "uuid_array", "multidim_array",
		}

		columnTypes := []string{
			"_text", "_integer", "_boolean", "_json",
			"_bytea", "_xml", "_uuid", "_integer[]",
		}
		result, err := parsePgRowValues(values, columnNames, columnTypes)
		require.NoError(t, err)

		expected := map[string]any{
			"text_array": []any{"Hello", "World"},
			"int_array":  []any{int64(42), int64(43)},
			"bool_array": []any{true, false},
			"json_array": []any{map[string]any{"key": "value1"}, map[string]any{"key": "value2"}},
			"binary_array": &neosynctypes.NeosyncArray{
				BaseType: neosynctypes.BaseType{
					Neosync: neosynctypes.Neosync{
						Version: 1,
						TypeId:  "NEOSYNC_ARRAY",
					},
				},
				Elements: []neosynctypes.NeosyncAdapter{
					&neosynctypes.Binary{
						BaseType: neosynctypes.BaseType{
							Neosync: neosynctypes.Neosync{
								Version: 1,
								TypeId:  "NEOSYNC_BINARY",
							},
						},
						Bytes: binaryData1,
					},
					&neosynctypes.Binary{
						BaseType: neosynctypes.BaseType{
							Neosync: neosynctypes.Neosync{
								Version: 1,
								TypeId:  "NEOSYNC_BINARY",
							},
						},
						Bytes: binaryData2,
					},
				},
			},
			"xml_array":      []any{xmlData1, xmlData2},
			"uuid_array":     []any{uuidValue1, uuidValue2},
			"multidim_array": []any{[]any{1, 2}, []any{3, 4}},
		}

		require.Equal(t, expected, result)
	})

	t.Run("Null Values", func(t *testing.T) {
		values := []any{
			&sql.NullString{Valid: false},
			&NullableJSON{Valid: false},
		}
		columnNames := []string{"null_string", "null_json"}
		columnTypes := []string{"text", "text"}

		result, err := parsePgRowValues(values, columnNames, columnTypes)
		require.NoError(t, err)

		expected := map[string]any{
			"null_string": nil,
			"null_json":   nil,
		}
		require.Equal(t, expected, result)
	})
}

func TestNullableJSON_Unmarshal(t *testing.T) {
	tests := []struct {
		name    string
		json    NullableJSON
		want    any
		wantErr bool
	}{
		{
			name:    "Invalid JSON",
			json:    NullableJSON{Valid: false},
			want:    nil,
			wantErr: false,
		},
		{
			name:    "Valid string",
			json:    NullableJSON{RawMessage: json.RawMessage(`"test"`), Valid: true},
			want:    "test",
			wantErr: false,
		},
		{
			name:    "Valid number",
			json:    NullableJSON{RawMessage: json.RawMessage(`42`), Valid: true},
			want:    float64(42),
			wantErr: false,
		},
		{
			name:    "Valid object",
			json:    NullableJSON{RawMessage: json.RawMessage(`{"key":"value"}`), Valid: true},
			want:    map[string]any{"key": "value"},
			wantErr: false,
		},
		{
			name:    "Invalid JSON content",
			json:    NullableJSON{RawMessage: json.RawMessage(`{invalid}`), Valid: true},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.json.Unmarshal()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_CreateMultiDimSlice(t *testing.T) {
	t.Run("1D Slice of Ints", func(t *testing.T) {
		dims := []int{3}
		elements := []any{1, 2, 3}
		result := createMultiDimSlice(dims, elements)
		expected := []any{1, 2, 3}
		require.Equal(t, expected, result)
	})

	t.Run("2D Slice of Ints", func(t *testing.T) {
		dims := []int{2, 3}
		elements := []any{1, 2, 3, 4, 5, 6}
		result := createMultiDimSlice(dims, elements)
		expected := [][]any{{1, 2, 3}, {4, 5, 6}}
		require.Equal(t, expected, result)
	})

	t.Run("3D Slice of Ints", func(t *testing.T) {
		dims := []int{2, 2, 2}
		elements := []any{1, 2, 3, 4, 5, 6, 7, 8}
		result := createMultiDimSlice(dims, elements)
		expected := [][][]any{{{1, 2}, {3, 4}}, {{5, 6}, {7, 8}}}
		require.Equal(t, expected, result)
	})

	t.Run("2D Slice of Strings", func(t *testing.T) {
		dims := []int{2, 2}
		elements := []any{"a", "b", "c", "d"}
		result := createMultiDimSlice(dims, elements)
		expected := [][]any{{"a", "b"}, {"c", "d"}}
		require.Equal(t, expected, result)
	})

	t.Run("Empty Dims", func(t *testing.T) {
		dims := []int{}
		elements := []any{42}
		result := createMultiDimSlice(dims, elements)
		require.Equal(t, []any{42}, result)
	})

	t.Run("1D Slice with Single Element", func(t *testing.T) {
		dims := []int{1}
		elements := []any{42}
		result := createMultiDimSlice(dims, elements)
		expected := []any{42}
		require.Equal(t, expected, result)
	})

	t.Run("3D Slice with Mixed Types", func(t *testing.T) {
		dims := []int{2, 2, 2}
		elements := []any{1, "a", true, 3.14, 0, 'b', []int{1, 2}, map[string]int{"x": 1}}
		result := createMultiDimSlice(dims, elements)
		expected := [][][]any{
			{{1, "a"}, {true, 3.14}},
			{{0, 'b'}, {[]int{1, 2}, map[string]int{"x": 1}}},
		}
		require.Equal(t, expected, result)
	})
}
