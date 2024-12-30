package sqlserver

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	mssql "github.com/microsoft/go-mssqldb"
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

	columnDbTypes := []string{}
	for _, c := range cTypes {
		columnDbTypes = append(columnDbTypes, c.DatabaseTypeName())
	}

	values := make([]any, len(columnNames))
	valuesWrapped := make([]any, 0, len(columnNames))
	for i := range values {
		colType := columnDbTypes[i]
		if strings.EqualFold(colType, "uniqueidentifier") {
			values[i] = &mssql.UniqueIdentifier{}
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
		colType := columnDbTypes[i]
		switch t := v.(type) {
		case time.Time:
			dt, err := neosynctypes.NewDateTimeFromMssql(t)
			if err != nil {
				jObj[col] = t
				continue
			}
			jObj[col] = dt
		case *mssql.UniqueIdentifier:
			jObj[col] = t.String()
		case []byte:
			switch {
			case strings.EqualFold(colType, "binary"):
				binary, err := neosynctypes.NewBinaryFromMssql(t)
				if err != nil {
					jObj[col] = t
					continue
				}
				jObj[col] = binary
			case strings.EqualFold(colType, "varbinary"):
				bits, err := neosynctypes.NewBitsFromMssql(t)
				if err != nil {
					jObj[col] = t
					continue
				}
				jObj[col] = bits
			default:
				jObj[col] = string(t)
			}
		default:
			jObj[col] = t
		}
	}

	jsonF, _ := json.MarshalIndent(jObj, "", " ")
	fmt.Printf("\n\n %s \n\n", string(jsonF))

	return jObj, nil
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
