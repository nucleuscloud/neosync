package databaserecordmapper

import (
	"database/sql"
	"errors"

	neosync_types "github.com/nucleuscloud/neosync/internal/types"
)

type MySQLMapper struct{}

func NewMySQLBuilder() *Builder[*sql.Rows] {
	return &Builder[*sql.Rows]{
		mapper: &MySQLMapper{},
	}
}

func (m *MySQLMapper) MapRecord(rows *sql.Rows) (map[string]any, error) {
	return nil, nil
}

func (m *MySQLMapper) MapRecordWithKeyType(rows *sql.Rows) (map[string]any, map[string]neosync_types.KeyType, error) {
	return nil, nil, errors.ErrUnsupported
}
