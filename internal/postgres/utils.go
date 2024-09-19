package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	gotypeutil "github.com/nucleuscloud/neosync/internal/gotypeutil"
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
		fmt.Printf("%s %s %T %+v \n\n", col, ctype.DatabaseTypeName(), v, v)
		switch t := v.(type) {
		case []byte:
			if isJsonPgDataType(ctype.DatabaseTypeName()) {
				jmap, err := gotypeutil.JsonToMap(t)
				if err == nil {
					jObj[col] = jmap
					continue
				}
			}
			jObj[col] = t
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

func isJsonArrayPgDataType(dataType string) bool {
	return strings.EqualFold(dataType, "_json") || strings.EqualFold(dataType, "_jsonb")
}

func pgArrayToGoSlice(array *PgxArray[any]) any {
	goSlice := convertArrayToGoType(array)

	dim := array.Dimensions()
	if len(dim) > 1 {
		dims := []int{}
		for _, d := range dim {
			dims = append(dims, int(d.Length))
		}
		return CreateMultiDimSlice(dims, goSlice)
	}
	return goSlice
}

func convertArrayToGoType(array *PgxArray[any]) []any {
	if !isJsonArrayPgDataType(array.colDataType) {
		if isPgUuidArray(array.colDataType) {
			return convertBytesToUuidSlice(array.Elements)
		}
		return array.Elements
	}

	var newArray []any
	for _, e := range array.Elements {
		jsonBits, ok := e.([]byte)
		if !ok {
			newArray = append(newArray, e)
			continue
		}

		jmap, err := gotypeutil.JsonToMap(jsonBits)
		if err != nil {
			newArray = append(newArray, e)
		} else {
			newArray = append(newArray, jmap)
		}
	}

	return newArray
}

func isPgUuidArray(colDataType string) bool {
	return strings.EqualFold(colDataType, "_uuid")
}

func convertBytesToUuidSlice(uuids []any) []any {
	uuidSlice := []any{}
	for _, el := range uuids {
		if id, ok := el.([16]uint8); ok {
			uuidSlice = append(uuidSlice, uuid.UUID(id).String())
		}
	}
	return uuidSlice
}

// converts flat slice to multi-dimensional slice
func CreateMultiDimSlice(dims []int, elements []any) any {
	if len(elements) == 0 {
		return elements
	}
	if len(dims) == 0 || len(dims) == 1 {
		return elements
	}
	firstDim := dims[0]

	// creates nested any slice where depth = # of dimensions
	// 2 dimensions creates [][]any{}
	sliceType := reflect.TypeOf([]any{})
	for i := 0; i < len(dims)-1; i++ {
		sliceType = reflect.SliceOf(sliceType)
	}
	slice := reflect.MakeSlice(sliceType, firstDim, firstDim)

	// handles multi-dimensional slices
	subSize := 1
	for _, dim := range dims[1:] { //nolint:gosec
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
func FormatPgArrayLiteral(arr any, castType string) string {
	arrayLiteral := "ARRAY" + formatArrayLiteral(arr)
	if castType == "" {
		return arrayLiteral
	}

	return arrayLiteral + "::" + castType
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
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return fmt.Sprintf("%v", m)
	}

	return fmt.Sprintf("'%s'", string(jsonBytes))
}

func IsPgArrayType(dbTypeName string) bool {
	return strings.HasPrefix(dbTypeName, "_")
}

func IsPgArrayColumnDataType(colDataType string) bool {
	return strings.Contains(colDataType, "[]")
}
