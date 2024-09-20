package sqlserver

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/gofrs/uuid"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
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
	for col, val := range jObj {
		fmt.Printf("%s %T %+v \n\n", col, val, val)
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

func GeSqlServerDefaultValuesInsertSql(schema, table string, rowCount int) string {
	var sqlStr string
	for i := 0; i < rowCount; i++ {
		sqlStr += fmt.Sprintf("INSERT INTO %q.%q DEFAULT VALUES;", schema, table)
	}
	return sqlStr
}

func GoTypeToSqlServerType(rows [][]any) [][]any {
	newRows := [][]any{}
	for _, r := range rows {
		newRow := []any{}
		for _, v := range r {
			switch t := v.(type) {
			case bool:
				newRow = append(newRow, toBit(t))
			default:
				newRow = append(newRow, t)
			}
		}
		newRows = append(newRows, newRow)
	}
	return newRows
}

func toBit(v bool) int {
	if v {
		return 1
	}
	return 0
}

func FilterOutSqlServerDefaultIdentityColumns(
	driver string,
	identityCols, columnNames []string,
	argRows [][]any,
) (columns []string, rows [][]any) {
	if len(identityCols) == 0 || driver != sqlmanager_shared.MssqlDriver {
		return columnNames, argRows
	}

	// build map of identity columns
	identityColMap := map[string]bool{}
	for _, id := range identityCols {
		identityColMap[id] = true
	}

	nonIdentityColumnMap := map[string]struct{}{} // map of non identity columns
	newRows := [][]any{}
	// build rows removing identity columns/args with default set
	for _, row := range argRows {
		newRow := []any{}
		for idx, arg := range row {
			col := columnNames[idx]
			if identityColMap[col] && arg == "DEFAULT" {
				// pass on identity columns with a default
				continue
			}
			newRow = append(newRow, arg)
			nonIdentityColumnMap[col] = struct{}{}
		}
		newRows = append(newRows, newRow)
	}
	newColumns := []string{}
	// build new columns list while maintaining same order
	for _, col := range columnNames {
		if _, ok := nonIdentityColumnMap[col]; ok {
			newColumns = append(newColumns, col)
		}
	}
	return newColumns, newRows
}
