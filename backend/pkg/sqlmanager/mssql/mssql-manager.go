package sqlmanager_mssql

import (
	"context"
	"errors"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	mssql_queries "github.com/nucleuscloud/neosync/backend/pkg/mssql-querier"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
)

type Manager struct {
	querier mssql_queries.Querier
	db      mysql_queries.DBTX
	close   func()
}

func NewManager(querier mssql_queries.Querier, db mysql_queries.DBTX, closer func()) *Manager {
	return &Manager{querier: querier, db: db, close: closer}
}

func (m *Manager) GetDatabaseSchema(ctx context.Context) ([]*sqlmanager_shared.DatabaseSchemaRow, error) {
	return []*sqlmanager_shared.DatabaseSchemaRow{}, nil
}

func (m *Manager) GetSchemaColumnMap(ctx context.Context) (map[string]map[string]*sqlmanager_shared.ColumnInfo, error) {
	dbSchemas, err := m.GetDatabaseSchema(ctx)
	if err != nil {
		return nil, err
	}
	result := sqlmanager_shared.GetUniqueSchemaColMappings(dbSchemas)
	return result, nil
}

func (m *Manager) GetTableConstraintsBySchema(ctx context.Context, schemas []string) (*sqlmanager_shared.TableConstraints, error) {
	return &sqlmanager_shared.TableConstraints{
		ForeignKeyConstraints: map[string][]*sqlmanager_shared.ForeignConstraint{},
		PrimaryKeyConstraints: map[string][]string{},
		UniqueConstraints:     map[string][][]string{},
	}, nil
}

func (m *Manager) GetRolePermissionsMap(ctx context.Context) (map[string][]string, error) {
	rows, err := m.querier.GetRolePermissions(ctx, m.db)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return map[string][]string{}, nil
	}

	schemaTablePrivsMap := map[string][]string{}
	for _, permission := range rows {
		key := sqlmanager_shared.BuildTable(permission.TableSchema, permission.TableName)
		schemaTablePrivsMap[key] = append(schemaTablePrivsMap[key], permission.PrivilegeType)
	}
	return schemaTablePrivsMap, err
}

func (m *Manager) GetTableInitStatements(ctx context.Context, tables []*sqlmanager_shared.SchemaTable) ([]*sqlmanager_shared.TableInitStatement, error) {
	return []*sqlmanager_shared.TableInitStatement{}, nil
}

func (m *Manager) GetSchemaTableDataTypes(ctx context.Context, tables []*sqlmanager_shared.SchemaTable) (*sqlmanager_shared.SchemaTableDataTypeResponse, error) {
	return &sqlmanager_shared.SchemaTableDataTypeResponse{
		Sequences:  []*sqlmanager_shared.DataType{},
		Functions:  []*sqlmanager_shared.DataType{},
		Composites: []*sqlmanager_shared.DataType{},
		Enums:      []*sqlmanager_shared.DataType{},
		Domains:    []*sqlmanager_shared.DataType{},
	}, nil
}

func (m *Manager) GetSchemaTableTriggers(ctx context.Context, tables []*sqlmanager_shared.SchemaTable) ([]*sqlmanager_shared.TableTrigger, error) {
	return []*sqlmanager_shared.TableTrigger{}, nil
}

func (m *Manager) GetSchemaInitStatements(ctx context.Context, tables []*sqlmanager_shared.SchemaTable) ([]*sqlmanager_shared.InitSchemaStatements, error) {
	return []*sqlmanager_shared.InitSchemaStatements{}, nil
}

func (m *Manager) GetCreateTableStatement(ctx context.Context, schema, table string) (string, error) {
	return "", errors.ErrUnsupported
}

func (m *Manager) BatchExec(ctx context.Context, batchSize int, statements []string, opts *sqlmanager_shared.BatchExecOpts) error {
	return errors.ErrUnsupported
}

func (m *Manager) GetTableRowCount(
	ctx context.Context,
	schema, table string,
	whereClause *string,
) (int64, error) {
	return -1, errors.ErrUnsupported
}

func (m *Manager) Exec(ctx context.Context, statement string) error {
	_, err := m.db.ExecContext(ctx, statement)
	return err
}

func (m *Manager) Close() {
	if m.db != nil && m.close != nil {
		m.close()
	}
}
