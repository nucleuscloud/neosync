package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

func SqlRowToPgTypesMap(rows *sql.Rows) (map[string]any, error) {
	columnNames, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	cTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}
	colTypes := map[string]*sql.ColumnType{}
	for _, ct := range cTypes {
		colTypes[ct.Name()] = ct
	}

	values := make([]any, len(columnNames))
	valuesWrapped := make([]any, 0, len(columnNames))
	for i := range values {
		ctype := cTypes[i]
		if isPgArrayType(ctype.DatabaseTypeName()) {
			values[i] = &Array[any]{}
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
		// dbTypeName := colTypes[col].DatabaseTypeName()
		if col == "integer_array_col" {
			fmt.Printf("%s %+v %T \n", col, v, v)
			fmt.Printf("%+v \n", &valuesWrapped[i])
		}
		jObj[col] = v
		// if isPgArrayType(dbTypeName) {
		// 	jObj[col] = pq.Array([]byte(v))
		// } else {
		// 	jObj[col] = v
		// }
	}

	return jObj, nil
}

type Array[T any] struct {
	pgtype.Array[T]
}

func (a *Array[T]) Scan(src any) error {
	m := pgtype.NewMap()

	v := (*pgtype.Array[T])(&a.Array)

	t, ok := m.TypeForValue(v)
	if !ok {
		return fmt.Errorf("cannot convert to sql.Scanner: cannot find registered type for %T", a)
	}

	fmt.Println("HERE")
	fmt.Println(src, v, t)
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

	return m.Scan(t.OID, pgtype.TextFormatCode, bufSrc, v)
}

// type Array[T any] []T

// func (a *Array[T]) Scan(src any) error {
// 	m := pgtype.NewMap()

// 	v := (*[]T)(a)
// 	// fmt.Printf("%v+ \n", v, v)

// 	t, ok := m.TypeForValue(v)
// 	if !ok {
// 		return fmt.Errorf("cannot convert to sql.Scanner: cannot find registered type for %T", a)
// 	}

// 	var bufSrc []byte
// 	if src != nil {
// 		switch src := src.(type) {
// 		case string:
// 			bufSrc = []byte(src)
// 		case []byte:
// 			bufSrc = src
// 		default:
// 			bufSrc = []byte(fmt.Sprint(bufSrc))
// 		}
// 	}

// 	return m.Scan(t.OID, pgtype.TextFormatCode, bufSrc, v)
// }

func isPgArrayType(dbTypeName string) bool {
	return strings.HasPrefix(dbTypeName, "_")
}

// // convert '{1,2,3}' to []any{1,2,3}
// func ConvertPgArrayToSlice(input any) []any {
// 	strInput, ok := input.(string)
// 	if !ok || strInput == "" || strInput == "{}" {
// 		return []any{}
// 	}
// 	fmt.Println(strInput)

// 	strInput = strings.Trim(strInput, "{}")
// 	parts := strings.Split(strInput, ",")
// 	fmt.Println(parts)

// 	result := make([]any, len(parts))
// 	for i, part := range parts {
// 		fmt.Println(part)
// 		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
// 			result[i] = ConvertPgArrayToSlice(part)
// 		} else {
// 			result[i] = part
// 		}
// 	}

// 	return result
// }

// convert []any{1,2,3} to '{1,2,3}'
func ConvertSliceToPgString(s []any) string {
	var builder strings.Builder

	builder.WriteString("{")
	for i, v := range s {
		if i == len(s)-1 {
			builder.WriteString(fmt.Sprintf("%v", v))
		} else {
			builder.WriteString(fmt.Sprintf("%v,", v))
		}
	}
	builder.WriteString("}")
	return builder.String()
}

func ToPgTypes(args []any) []any {
	newArgs := []any{}
	for _, val := range args {
		switch v := val.(type) {
		case []any:
			newArgs = append(newArgs, ConvertSliceToPgString(v))
		default:
			newArgs = append(newArgs, v)
		}
	}
	return newArgs
}
