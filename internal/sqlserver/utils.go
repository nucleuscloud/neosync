package sqlserver

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	neosynctypes "github.com/nucleuscloud/neosync/internal/neosync-types"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
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
		case time.Time:
			dt, err := neosynctypes.NewDateTimeFromMssql(t)
			if err != nil {
				jObj[col] = t
				continue
			}
			jObj[col] = dt
		case []byte:
			if IsUuidDataType(colType.DatabaseTypeName()) {
				uuidStr, err := BitsToUuidString(t)
				if err == nil {
					jObj[col] = uuidStr
					continue
				}
			}
			jObj[col] = string(t)
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
	columnNames []string,
	argRows [][]any,
	colDefaultProperties []*neosync_benthos.ColumnDefaultProperties,
) (columns []string, rows [][]any, columnDefaultProperties []*neosync_benthos.ColumnDefaultProperties) {
	// build map of identity columns
	defaultIdentityColMap := map[string]bool{}
	for idx, d := range colDefaultProperties {
		cName := columnNames[idx]
		if d != nil && d.HasDefaultTransformer && d.NeedsOverride && d.NeedsReset {
			defaultIdentityColMap[cName] = true
		}
	}

	if driver != sqlmanager_shared.MssqlDriver || len(defaultIdentityColMap) == 0 {
		return columnNames, argRows, colDefaultProperties
	}

	nonIdentityColumnMap := map[string]struct{}{} // map of non identity columns
	newRows := [][]any{}
	// build rows removing identity columns/args with default set
	for _, row := range argRows {
		newRow := []any{}
		for idx, arg := range row {
			col := columnNames[idx]
			if defaultIdentityColMap[col] {
				// pass on identity columns with a default
				continue
			}
			newRow = append(newRow, arg)
			nonIdentityColumnMap[col] = struct{}{}
		}
		if len(newRow) != 0 {
			newRows = append(newRows, newRow)
		}
	}
	newColumns := []string{}
	newDefaultProperites := []*neosync_benthos.ColumnDefaultProperties{}
	// build new columns list while maintaining same order
	for idx, col := range columnNames {
		if _, ok := nonIdentityColumnMap[col]; ok {
			newColumns = append(newColumns, col)
			newDefaultProperites = append(newDefaultProperites, colDefaultProperties[idx])
		}
	}
	return newColumns, newRows, newDefaultProperites
}
