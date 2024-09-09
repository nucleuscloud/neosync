package sqlserver

import (
	"database/sql"
	"strings"

	"github.com/gofrs/uuid"
)

func SqlRowToSqlServerTypesMap(rows *sql.Rows) (map[string]any, error) {
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
		valuesWrapped = append(valuesWrapped, &values[i])
	}
	if err := rows.Scan(valuesWrapped...); err != nil {
		return nil, err
	}

	jObj := map[string]any{}
	for i, v := range values {
		col := columnNames[i]
		colType := cTypes[i]
		switch t := v.(type) {
		case string:
			jObj[col] = t
		case []byte:
			if IsUuidDataType(colType.DatabaseTypeName()) {
				uuidStr, err := BitsToUuidString(t)
				if err == nil {
					jObj[col] = uuidStr
					continue
				}
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

func IsUuidDataType(colDataType string) bool {
	return strings.EqualFold(colDataType, "uniqueidentifier")
}

func BitsToUuidString(bits []byte) (string, error) {
	u, err := uuid.FromBytes(bits)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}
