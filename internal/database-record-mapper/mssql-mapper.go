package databaserecordmapper

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	neosynctypes "github.com/nucleuscloud/neosync/internal/neosync-types"
	"github.com/nucleuscloud/neosync/internal/sqlserver"
	neosync_types "github.com/nucleuscloud/neosync/internal/types"
)

type MSSQLMapper struct{}

func NewMSSQLBuilder() *Builder[*sql.Rows] {
	return &Builder[*sql.Rows]{
		mapper: &MSSQLMapper{},
	}
}

func (m *MSSQLMapper) MapRecordWithKeyType(rows *sql.Rows) (map[string]any, map[string]neosync_types.KeyType, error) {
	return nil, nil, errors.ErrUnsupported
}

func (m *MSSQLMapper) MapRecord(rows *sql.Rows) (map[string]any, error) {
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
			if strings.EqualFold(colType.DatabaseTypeName(), "uniqueidentifier") {
				uuidStr, err := sqlserver.BitsToUuidString(t)
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
