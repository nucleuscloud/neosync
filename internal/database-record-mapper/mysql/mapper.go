package mysql

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nucleuscloud/neosync/internal/database-record-mapper/builder"
	neosynctypes "github.com/nucleuscloud/neosync/internal/neosync-types"
	neosync_types "github.com/nucleuscloud/neosync/internal/types"
)

type MySQLMapper struct{}

func NewMySQLBuilder() *builder.Builder[*sql.Rows] {
	return &builder.Builder[*sql.Rows]{
		Mapper: &MySQLMapper{},
	}
}

func (m *MySQLMapper) MapRecordWithKeyType(rows *sql.Rows) (valuemap map[string]any, typemap map[string]neosync_types.KeyType, err error) {
	return nil, nil, errors.ErrUnsupported
}

func (m *MySQLMapper) MapRecord(rows *sql.Rows) (map[string]any, error) {
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
	jObj, err := parseMysqlRowValues(values, columnNames, columnDbTypes)
	if err != nil {
		return nil, err
	}

	return jObj, nil
}

func parseMysqlRowValues(values []any, columnNames, columnDbTypes []string) (map[string]any, error) {
	jObj := map[string]any{}
	for i, v := range values {
		col := columnNames[i]
		colDataType := columnDbTypes[i]
		switch t := v.(type) {
		case time.Time:
			dt, err := neosynctypes.NewDateTimeFromMysql(t)
			if err != nil {
				return nil, fmt.Errorf("failed to parse datetime value: %w", err)
			}
			jObj[col] = dt
		case []byte:
			if strings.EqualFold(colDataType, "json") {
				var js any
				if err := json.Unmarshal(t, &js); err != nil {
					return nil, err
				}
				jObj[col] = js
			} else if strings.EqualFold(colDataType, "binary") {
				binary, err := neosynctypes.NewBinaryFromMysql(t)
				if err != nil {
					return nil, fmt.Errorf("failed to parse binary value: %w", err)
				}
				jObj[col] = binary
			} else if strings.EqualFold(colDataType, "bit") || strings.EqualFold(colDataType, "varbit") {
				bits, err := neosynctypes.NewBitsFromMysql(t)
				if err != nil {
					return nil, fmt.Errorf("failed to parse bit/varbit value: %w", err)
				}
				jObj[col] = bits
			} else {
				jObj[col] = string(t)
			}
		default:
			jObj[col] = t
		}
	}
	return jObj, nil
}
