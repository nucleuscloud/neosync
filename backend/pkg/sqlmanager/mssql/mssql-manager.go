package sqlmanager_mssql

import (
	"context"
	"errors"
	"fmt"

	"github.com/doug-martin/goqu/v9"
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
	dbSchemas, err := m.querier.GetDatabaseSchema(ctx, m.db)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return []*sqlmanager_shared.DatabaseSchemaRow{}, nil
	}

	output := []*sqlmanager_shared.DatabaseSchemaRow{}
	for _, row := range dbSchemas {
		charMaxLength := int32(-1)
		if row.CharacterMaximumLength.Valid {
			charMaxLength = row.CharacterMaximumLength.Int32
		}
		numericPrecision := int32(-1)
		if row.NumericPrecision.Valid {
			numericPrecision = int32(row.NumericPrecision.Int16)
		}
		numericScale := int32(-1)
		if row.NumericScale.Valid {
			numericScale = int32(row.NumericScale.Int16)
		}
		output = append(output, &sqlmanager_shared.DatabaseSchemaRow{
			TableSchema:            row.TableSchema,
			TableName:              row.TableName,
			ColumnName:             row.ColumnName,
			DataType:               row.DataType,
			ColumnDefault:          row.ColumnDefault, // todo: make sure this is valid for the other funcs
			IsNullable:             row.IsNullable,
			GeneratedType:          nil, // todo
			OrdinalPosition:        int16(row.OrdinalPosition),
			CharacterMaximumLength: charMaxLength,
			NumericPrecision:       numericPrecision,
			NumericScale:           numericScale,
			IdentityGeneration:     nil, // todo: will have to update the downstream logic for this
		})
	}

	return output, nil
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
	// mssql does not support batching statements
	total := len(statements)
	for idx, stmt := range statements {
		err := m.Exec(ctx, stmt)
		if err != nil {
			return fmt.Errorf("failed to execute batch statement %d/%d: %w", idx+1, total, err)
		}
	}
	return nil
}

func (m *Manager) GetTableRowCount(
	ctx context.Context,
	schema, table string,
	whereClause *string,
) (int64, error) {
	tableName := sqlmanager_shared.BuildTable(schema, table)
	builder := goqu.Dialect(sqlmanager_shared.MssqlDriver)

	query := builder.From(goqu.I(tableName)).Select(goqu.COUNT("*"))
	if whereClause != nil && *whereClause != "" {
		query = query.Where(goqu.L(*whereClause))
	}
	sql, _, err := query.ToSQL()
	if err != nil {
		return 0, fmt.Errorf("unable to build table row count statement for mssql: %w", err)
	}
	var count int64
	err = m.db.QueryRowContext(ctx, sql).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("unable to query table row count for mssql: %w", err)
	}
	return count, err
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
