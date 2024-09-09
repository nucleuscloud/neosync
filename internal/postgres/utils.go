package postgres

import (
	"database/sql"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	gotypeparser "github.com/nucleuscloud/neosync/internal/gotype-parser"
)

type PgxArray[T any] struct {
	pgtype.Array[T]
	colDataType string
}

// custom PGX array scanner
// properly handles scanning postgres arrays
func (a *PgxArray[T]) Scan(src any) error {
	m := pgtype.NewMap()
	pgt, ok := m.TypeForName(strings.ToLower(a.colDataType))
	if !ok {
		return fmt.Errorf("cannot convert to sql.Scanner: cannot find registered type for %s", a.colDataType)
	}

	v := &a.Array
	var bufSrc []byte
	if src != nil {
		switch src := src.(type) {
		case string:
			bufSrc = []byte(src)
		case []byte:
			bufSrc = src
		default:
			bufSrc = []byte(fmt.Sprint(bufSrc))
		}
	}

	return m.Scan(pgt.OID, pgtype.TextFormatCode, bufSrc, v)
}

func SqlRowToPgTypesMap(rows *sql.Rows) (map[string]any, error) {
	columnNames, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	cTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	values := make([]any, len(columnNames))
	valuesWrapped := make([]any, 0, len(columnNames))
	for i := range values {
		ctype := cTypes[i]
		if IsPgArrayType(ctype.DatabaseTypeName()) {
			// use custom array type scanner
			values[i] = &PgxArray[any]{
				colDataType: ctype.DatabaseTypeName(),
			}
			valuesWrapped = append(valuesWrapped, values[i])
		} else {
			valuesWrapped = append(valuesWrapped, &values[i])
		}
	}
	if err := rows.Scan(valuesWrapped...); err != nil {
		return nil, err
	}

	jObj := map[string]any{}
	for i, v := range values {
		col := columnNames[i]
		ctype := cTypes[i]
		switch t := v.(type) {
		case []byte:
			if isJsonPgDataType(ctype.DatabaseTypeName()) {
				jmap, err := gotypeparser.JsonToMap(t)
				if err != nil {
					jObj[col] = string(t)
					continue
				}
				jObj[col] = jmap
				continue
			}
			jObj[col] = string(t)
		case *PgxArray[any]:
			jObj[col] = pgArrayToGoSlice(t)
		default:
			jObj[col] = t
		}
	}

	return jObj, nil
}

func isJsonPgDataType(dataType string) bool {
	return strings.EqualFold(dataType, "json") || strings.EqualFold(dataType, "jsonb")
}

func GoTypeToPgType(rows [][]any) [][]any {
	newRows := [][]any{}
	for _, r := range rows {
		newRow := []any{}
		for _, v := range r {
			switch t := v.(type) {
			case map[string]any:
				bits, err := gotypeparser.MapToJson(t)
				if err != nil {
					newRow = append(newRow, t)
					continue
				}
				newRow = append(newRow, bits)
			default:
				newRow = append(newRow, t)
			}
		}
		newRows = append(newRows, newRow)
	}
	return newRows
}

func pgArrayToGoSlice(array *PgxArray[any]) any {
	dim := array.Dimensions()
	if len(dim) > 1 {
		dims := []int{}
		for _, d := range dim {
			dims = append(dims, int(d.Length))
		}
		return CreateMultiDimSlice(dims, array.Elements)
	}
	return array.Elements
}

func IsMultiDimensionalSlice(val any) bool {
	rv := reflect.ValueOf(val)

	if rv.Kind() != reflect.Slice {
		return false
	}

	// if the slice is empty can't determine if it's multi-dimensional
	if rv.Len() == 0 {
		return false
	}

	firstElem := rv.Index(0)
	// if an interface check underlying value
	if firstElem.Kind() == reflect.Interface {
		firstElem = firstElem.Elem()
	}

	return firstElem.Kind() == reflect.Slice
}

// converts flat slice to multi-dimensional slice
func CreateMultiDimSlice(dims []int, elements []any) any {
	if len(elements) == 0 {
		return elements
	}
	if len(dims) == 0 {
		return elements[0]
	}
	firstDim := dims[0]

	// creates nested any slice where depth = # of dimensions
	// 2 dimensions creates [][]any{}
	sliceType := reflect.TypeOf([]any{})
	for i := 0; i < len(dims)-1; i++ {
		sliceType = reflect.SliceOf(sliceType)
	}
	slice := reflect.MakeSlice(sliceType, firstDim, firstDim)

	// handles 1 dimension slice []any{}
	if len(dims) == 1 {
		for i := 0; i < firstDim; i++ {
			slice.Index(i).Set(reflect.ValueOf(elements[i]))
		}
		return slice.Interface()
	}

	// handles multi-dimensional slices
	subSize := 1
	for _, dim := range dims[1:] {
		subSize *= dim
	}

	for i := 0; i < firstDim; i++ {
		start := i * subSize
		end := start + subSize
		subSlice := CreateMultiDimSlice(dims[1:], elements[start:end]) //nolint:gosec
		slice.Index(i).Set(reflect.ValueOf(subSlice))
	}

	return slice.Interface()
}

// returns string in this form ARRAY[[a,b],[c,d]]
func FormatPgArrayLiteral(arr any) string {
	return "ARRAY" + formatArrayLiteral(arr)
}

func formatArrayLiteral(arr any) string {
	v := reflect.ValueOf(arr)

	if v.Kind() == reflect.Slice {
		result := "["
		for i := 0; i < v.Len(); i++ {
			if i > 0 {
				result += ","
			}
			result += formatArrayLiteral(v.Index(i).Interface())
		}
		result += "]"
		return result
	}

	switch val := arr.(type) {
	case map[string]any:
		return formatMapLiteral(val)
	case string:
		return fmt.Sprintf("'%s'", strings.ReplaceAll(val, "'", "''"))
	default:
		return fmt.Sprintf("%v", val)
	}
}

func formatMapLiteral(m map[string]any) string {
	pairs := make([]string, 0, len(m))
	for k, v := range m {
		pairs = append(pairs, fmt.Sprintf("('%s',%s)", strings.ReplaceAll(k, "'", "''"), formatArrayLiteral(v)))
	}
	sort.Strings(pairs) // sort for consistent output
	return strings.Join(pairs, ",")
}

func IsPgArrayType(dbTypeName string) bool {
	return strings.HasPrefix(dbTypeName, "_")
}
