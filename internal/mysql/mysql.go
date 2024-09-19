package mysql

import (
	"database/sql"
	"strings"

	"github.com/nucleuscloud/neosync/internal/gotypeutil"
	"github.com/nucleuscloud/neosync/internal/sqlscanners"
)

func MysqlSqlRowToMap(rows *sql.Rows) (map[string]any, error) {
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
		colType := cTypes[i].DatabaseTypeName()
		if isBitDataType(colType) {
			values[i] = &sqlscanners.BitString{}
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
		colDataType := cTypes[i].DatabaseTypeName()
		switch t := v.(type) {
		case string:
			jObj[col] = t
		case []byte:
			if isJsonDataType(colDataType) {
				jmap, err := gotypeutil.JsonToMap(t)
				if err == nil {
					jObj[col] = jmap
				}
				continue
			}
			if isBinaryDataType(colDataType) {
				jObj[col] = t
			} else {
				jObj[col] = string(t)
			}
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			jObj[col] = t
		case float32, float64:
			jObj[col] = t
		case bool:
			jObj[col] = t
		default:
			jObj[col] = t
		}
	}

	return jObj, nil
}

func isBitDataType(colDataType string) bool {
	return strings.EqualFold(colDataType, "bit")
}

func isBinaryDataType(colDataType string) bool {
	return strings.EqualFold(colDataType, "binary")
}

func isJsonDataType(colDataType string) bool {
	return strings.EqualFold(colDataType, "json")
}
