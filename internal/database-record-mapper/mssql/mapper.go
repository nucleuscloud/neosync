package mssql

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	mssql "github.com/microsoft/go-mssqldb"
	"github.com/nucleuscloud/neosync/internal/database-record-mapper/builder"
	neosynctypes "github.com/nucleuscloud/neosync/internal/neosync-types"
	neosync_types "github.com/nucleuscloud/neosync/internal/types"
)

type MSSQLMapper struct{}

func NewMSSQLBuilder() *builder.Builder[*sql.Rows] {
	return &builder.Builder[*sql.Rows]{
		Mapper: &MSSQLMapper{},
	}
}

func (m *MSSQLMapper) MapRecordWithKeyType(
	rows *sql.Rows,
) (valuemap map[string]any, typemap map[string]neosync_types.KeyType, err error) {
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

	return parseRowValues(values, columnNames, columnDbTypes)
}

func parseRowValues(values []any, columnNames, columnDbTypes []string) (map[string]any, error) {
	jObj := map[string]any{}
	for i, v := range values {
		col := columnNames[i]
		colType := columnDbTypes[i]
		switch t := v.(type) {
		case time.Time:
			dt, err := neosynctypes.NewDateTimeFromMssql(t)
			if err != nil {
				return nil, fmt.Errorf("failed to convert time.Time to DateTime for column %s: %w", col, err)
			}
			jObj[col] = dt
		case *mssql.UniqueIdentifier:
			jObj[col] = t.String()
		case []byte:
			switch {
			case strings.EqualFold(colType, "binary"):
				binary, err := neosynctypes.NewBinaryFromMssql(t)
				if err != nil {
					return nil, fmt.Errorf("failed to convert binary data for column %s: %w", col, err)
				}
				jObj[col] = binary
			case strings.EqualFold(colType, "varbinary"):
				bits, err := neosynctypes.NewBitsFromMssql(t)
				if err != nil {
					return nil, fmt.Errorf("failed to convert varbinary data for column %s: %w", col, err)
				}
				jObj[col] = bits
			default:
				jObj[col] = string(t)
			}
		default:
			jObj[col] = t
		}
	}
	return jObj, nil
}
