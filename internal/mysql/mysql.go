package mysql

import (
	"database/sql"
	"encoding/json"
	"strings"

	neosynctypes "github.com/nucleuscloud/neosync/internal/neosync-types"
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
	columnDbTypes := []string{}
	for _, c := range cTypes {
		columnDbTypes = append(columnDbTypes, c.DatabaseTypeName())
	}

	values := make([]any, len(columnNames))
	valuesWrapped := make([]any, 0, len(columnNames))
	for i := range values {
		// colType := cTypes[i].DatabaseTypeName()
		// if isBitDataType(colType) {
		// 	values[i] = &sqlscanners.BitString{}
		// 	valuesWrapped = append(valuesWrapped, values[i])
		// } else {
		// 	valuesWrapped = append(valuesWrapped, &values[i])
		// }
		valuesWrapped = append(valuesWrapped, &values[i])
	}
	if err := rows.Scan(valuesWrapped...); err != nil {
		return nil, err
	}
	jObj := parseMysqlRowValues(values, columnNames, columnDbTypes)

	return jObj, nil
}

func parseMysqlRowValues(values []any, columnNames, columnDbTypes []string) map[string]any {
	jObj := map[string]any{}
	for i, v := range values {
		col := columnNames[i]
		colDataType := columnDbTypes[i]
		switch t := v.(type) {
		case string:
			jObj[col] = t
		case []byte:
			if IsJsonDataType(colDataType) {
				var js any
				if err := json.Unmarshal(t, &js); err == nil {
					jObj[col] = js
					continue
				}
			} else if isBinaryDataType(colDataType) {
				binary, err := neosynctypes.NewBinaryFromMysql(t)
				if err != nil {
					jObj[col] = t
					continue
				}
				jObj[col] = binary
			} else if isBitDataType(colDataType) || strings.EqualFold(colDataType, "varbit") {
				bits, err := neosynctypes.NewBitsFromMysql(t)
				if err != nil {
					jObj[col] = t
					continue
				}
				jObj[col] = bits
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
	return jObj
}

func isBitDataType(colDataType string) bool {
	return strings.EqualFold(colDataType, "bit")
}

func isBinaryDataType(colDataType string) bool {
	return strings.EqualFold(colDataType, "binary")
}

func IsJsonDataType(colDataType string) bool {
	return strings.EqualFold(colDataType, "json")
}
