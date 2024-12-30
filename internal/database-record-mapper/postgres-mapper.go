package databaserecordmapper

import (
	"database/sql"
	"errors"

	neosync_types "github.com/nucleuscloud/neosync/internal/types"
)

type PostgresMapper struct{}

func NewPostgresBuilder() *Builder[*sql.Rows] {
	return &Builder[*sql.Rows]{
		mapper: &PostgresMapper{},
	}
}

func (m *PostgresMapper) MapRecordWithKeyType(rows *sql.Rows) (map[string]any, map[string]neosync_types.KeyType, error) {
	return nil, nil, errors.ErrUnsupported
}

func (m *PostgresMapper) MapRecord(rows *sql.Rows) (map[string]any, error) {
	return nil, nil
}
