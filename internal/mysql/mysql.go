package mysql

import (
	"database/sql"
	"strings"

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
			var bitStr sqlscanners.BitString
			values[i] = &bitStr
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
		colDataType := cTypes[i]
		switch t := v.(type) {
		case string:
			jObj[col] = t
		case []byte:
			if strings.EqualFold(colDataType.DatabaseTypeName(), "binary") {
				jObj[col] = t
				continue
			}
			jObj[col] = string(t)
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
