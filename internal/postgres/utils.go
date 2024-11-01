package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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

	columnDbTypes := []string{}
	for _, c := range cTypes {
		columnDbTypes = append(columnDbTypes, c.DatabaseTypeName())
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

	jObj := parsePgRowValues(values, columnNames, columnDbTypes)
	return jObj, nil
}

func parsePgRowValues(values []any, columnNames, columnDbTypes []string) map[string]any {
	jObj := map[string]any{}
	for i, v := range values {
		col := columnNames[i]
		ctype := columnDbTypes[i]
		switch t := v.(type) {
		case []byte:
			if IsJsonPgDataType(ctype) {
				var js any
				if err := json.Unmarshal(t, &js); err == nil {
					jObj[col] = js
					continue
				}
			} else if isBinaryDataType(ctype) {
				jObj[col] = t
				continue
			}
			jObj[col] = string(t)
		case *PgxArray[any]:
			jObj[col] = pgArrayToGoSlice(t)
		default:
			jObj[col] = t
		}
	}
	return jObj
}

func isBinaryDataType(colDataType string) bool {
	return strings.EqualFold(colDataType, "bytea")
}

func IsJsonPgDataType(dataType string) bool {
	return strings.EqualFold(dataType, "json") || strings.EqualFold(dataType, "jsonb")
}

func isJsonArrayPgDataType(dataType string) bool {
	return strings.EqualFold(dataType, "_json") || strings.EqualFold(dataType, "_jsonb")
}

func isPgUuidArray(colDataType string) bool {
	return strings.EqualFold(colDataType, "_uuid")
}

func isPgXmlArray(colDataType string) bool {
	return strings.EqualFold(colDataType, "_xml")
}

func IsPgArrayType(dbTypeName string) bool {
	return strings.HasPrefix(dbTypeName, "_")
}

func IsPgArrayColumnDataType(colDataType string) bool {
	return strings.Contains(colDataType, "[]")
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
		if isPgXmlArray(array.colDataType) {
			return convertBytesToStringSlice(array.Elements)
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

		var js any
		err := json.Unmarshal(jsonBits, &js)
		if err != nil {
			newArray = append(newArray, e)
		} else {
			newArray = append(newArray, js)
		}
	}

	return newArray
}

func convertBytesToStringSlice(bytes []any) []any {
	stringSlice := []any{}
	for _, el := range bytes {
		if bits, ok := el.([]byte); ok {
			stringSlice = append(stringSlice, string(bits))
		}
	}
	return stringSlice
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
	for _, dim := range dims[1:] {
		subSize *= dim
	}

	for i := 0; i < firstDim; i++ {
		start := i * subSize
		end := start + subSize
		subSlice := CreateMultiDimSlice(dims[1:], elements[start:end])
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
