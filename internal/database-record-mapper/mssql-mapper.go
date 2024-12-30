package databaserecordmapper

import (
	"database/sql"
	"errors"

	neosync_types "github.com/nucleuscloud/neosync/internal/types"
)

type MSSQLMapper struct{}

func NewMSSQLBuilder() *Builder[*sql.Rows] {
	return &Builder[*sql.Rows]{
		mapper: &MSSQLMapper{},
	}
}

func (m *MSSQLMapper) MapRecord(rows *sql.Rows) (map[string]any, error) {
	return nil, nil
}

func (m *MSSQLMapper) MapRecordWithKeyType(rows *sql.Rows) (map[string]any, map[string]neosync_types.KeyType, error) {
	return nil, nil, errors.ErrUnsupported
}
