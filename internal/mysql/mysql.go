package mysql

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"

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
		case time.Time:
			dt, err := neosynctypes.NewDateTimeFromMysql(t)
			if err != nil {
				jObj[col] = t
				continue
			}
			jObj[col] = dt
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
