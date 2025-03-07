package connectiondata

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	sqlmanager_mysql "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/mysql"
	sqlmanager_postgres "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/postgres"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	database_record_mapper "github.com/nucleuscloud/neosync/internal/database-record-mapper"
	nucleuserrors "github.com/nucleuscloud/neosync/internal/errors"
	"github.com/nucleuscloud/neosync/internal/neosyncdb"
	querybuilder "github.com/nucleuscloud/neosync/worker/pkg/query-builder"
)

type SQLConnectionDataService struct {
	logger       *slog.Logger
	sqlconnector sqlconnect.SqlConnector
	sqlmanager   sql_manager.SqlManagerClient
	connection   *mgmtv1alpha1.Connection
	connconfig   *mgmtv1alpha1.ConnectionConfig
}

func NewSQLConnectionDataService(
	logger *slog.Logger,
	sqlconnector sqlconnect.SqlConnector,
	sqlmanager sql_manager.SqlManagerClient,
	connection *mgmtv1alpha1.Connection,
) *SQLConnectionDataService {
	return &SQLConnectionDataService{
		logger:       logger,
		sqlconnector: sqlconnector,
		sqlmanager:   sqlmanager,
		connection:   connection,
		connconfig:   connection.GetConnectionConfig(),
	}
}

func (s *SQLConnectionDataService) SampleData(
	ctx context.Context,
	stream SampleDataStream,
	schema, table string,
	numRows uint,
) error {
	err := s.areSchemaAndTableValid(ctx, schema, table)
	if err != nil {
		return fmt.Errorf("invalid schema or table: %w", err)
	}

	conn, err := s.sqlconnector.NewDbFromConnectionConfig(s.connconfig, s.logger, sqlconnect.WithConnectionTimeout(uint32(5)))
	if err != nil {
		return fmt.Errorf("error creating connection: %w", err)
	}
	defer conn.Close()
	db, err := conn.Open()
	if err != nil {
		return fmt.Errorf("error opening connection: %w", err)
	}

	mapper, err := database_record_mapper.NewDatabaseRecordMapperFromConnection(s.connection)
	if err != nil {
		return err
	}
	goquDriver, err := querybuilder.GetGoquDriverFromConnection(s.connection)
	if err != nil {
		return err
	}

	schemaTable := sqlmanager_shared.BuildTable(schema, table)

	query, err := querybuilder.BuildSampledSelectLimitQuery(goquDriver, schemaTable, numRows)
	if err != nil {
		return err
	}
	rows, err := db.QueryContext(ctx, query)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return fmt.Errorf("error querying table %s with database type %s: %w", schemaTable, goquDriver, err)
	}
	defer rows.Close()

	for rows.Next() {
		r, err := mapper.MapRecord(rows)
		if err != nil {
			return fmt.Errorf("unable to convert row to map for table %s with database type %s: %w", schemaTable, goquDriver, err)
		}
		var rowbytes bytes.Buffer
		enc := gob.NewEncoder(&rowbytes)
		if err := enc.Encode(r); err != nil {
			return fmt.Errorf("unable to encode row for table %s with database type %s: %w", schemaTable, goquDriver, err)
		}
		if err := stream.Send(&mgmtv1alpha1.GetConnectionDataStreamResponse{RowBytes: rowbytes.Bytes()}); err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLConnectionDataService) StreamData(
	ctx context.Context,
	stream *connect.ServerStream[mgmtv1alpha1.GetConnectionDataStreamResponse],
	config *mgmtv1alpha1.ConnectionStreamConfig,
	schema, table string,
) error {
	err := s.areSchemaAndTableValid(ctx, schema, table)
	if err != nil {
		return err
	}

	conn, err := s.sqlconnector.NewDbFromConnectionConfig(s.connconfig, s.logger, sqlconnect.WithConnectionTimeout(uint32(5)))
	if err != nil {
		return err
	}
	defer conn.Close()
	db, err := conn.Open()
	if err != nil {
		return err
	}

	mapper, err := database_record_mapper.NewDatabaseRecordMapperFromConnection(s.connection)
	if err != nil {
		return err
	}

	goquDriver, err := querybuilder.GetGoquDriverFromConnection(s.connection)
	if err != nil {
		return err
	}

	schemaTable := sqlmanager_shared.BuildTable(schema, table)
	// used to get column names
	query, err := querybuilder.BuildSelectLimitQuery(goquDriver, schemaTable, 0)
	if err != nil {
		return err
	}
	r, err := db.QueryContext(ctx, query)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return fmt.Errorf("error querying table %s with database type %s: %w", schemaTable, goquDriver, err)
	}

	columnNames, err := r.Columns()
	if err != nil {
		return fmt.Errorf("unable to get column names from table %s with database type %s: %w", schemaTable, goquDriver, err)
	}

	selectQuery, err := querybuilder.BuildSelectQuery(goquDriver, schemaTable, columnNames, nil)
	if err != nil {
		return err
	}
	rows, err := db.QueryContext(ctx, selectQuery)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return fmt.Errorf("error querying table %s with goqu driver %s: %w", schemaTable, goquDriver, err)
	}

	// todo: rows.Close needs to be called here?
	for rows.Next() {
		r, err := mapper.MapRecord(rows)
		if err != nil {
			return fmt.Errorf("unable to convert row to map for table %s with database type %s: %w", schemaTable, goquDriver, err)
		}
		var rowbytes bytes.Buffer
		enc := gob.NewEncoder(&rowbytes)
		if err := enc.Encode(r); err != nil {
			return fmt.Errorf("unable to encode row for table %s with database type %s: %w", schemaTable, goquDriver, err)
		}
		if err := stream.Send(&mgmtv1alpha1.GetConnectionDataStreamResponse{RowBytes: rowbytes.Bytes()}); err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLConnectionDataService) GetSchema(ctx context.Context, config *mgmtv1alpha1.ConnectionSchemaConfig) ([]*mgmtv1alpha1.DatabaseColumn, error) {
	db, err := s.sqlmanager.NewSqlConnection(ctx, connectionmanager.NewUniqueSession(), s.connection, s.logger)
	if err != nil {
		return nil, err
	}
	defer db.Db().Close()

	dbschema, err := db.Db().GetDatabaseSchema(ctx)
	if err != nil {
		return nil, err
	}
	schemas := []*mgmtv1alpha1.DatabaseColumn{}
	for _, col := range dbschema {
		col := col
		var defaultColumn *string
		if col.ColumnDefault != "" {
			defaultColumn = &col.ColumnDefault
		}

		schemas = append(schemas, &mgmtv1alpha1.DatabaseColumn{
			Schema:             col.TableSchema,
			Table:              col.TableName,
			Column:             col.ColumnName,
			DataType:           col.DataType,
			IsNullable:         col.NullableString(),
			ColumnDefault:      defaultColumn,
			GeneratedType:      col.GeneratedType,
			IdentityGeneration: col.IdentityGeneration,
		})
	}
	return schemas, nil
}

func (s *SQLConnectionDataService) GetInitStatements(
	ctx context.Context,
	options *mgmtv1alpha1.InitStatementOptions,
) (*mgmtv1alpha1.GetConnectionInitStatementsResponse, error) {
	schemas, err := s.GetSchema(ctx, nil)
	if err != nil {
		return nil, err
	}

	schemaTableMap := map[string]*mgmtv1alpha1.DatabaseColumn{}
	for _, s := range schemas {
		schemaTableMap[sqlmanager_shared.BuildTable(s.Schema, s.Table)] = s
	}

	db, err := s.sqlmanager.NewSqlConnection(ctx, connectionmanager.NewUniqueSession(), s.connection, s.logger)
	if err != nil {
		return nil, err
	}
	defer db.Db().Close()

	truncateStmtsMap := map[string]string{}
	initSchemaStmts := []*mgmtv1alpha1.SchemaInitStatements{}
	if options.GetInitSchema() {
		tables := []*sqlmanager_shared.SchemaTable{}
		for _, v := range schemaTableMap {
			tables = append(tables, &sqlmanager_shared.SchemaTable{Schema: v.Schema, Table: v.Table})
		}
		initBlocks, err := db.Db().GetSchemaInitStatements(ctx, tables)
		if err != nil {
			return nil, err
		}
		for _, b := range initBlocks {
			initSchemaStmts = append(initSchemaStmts, &mgmtv1alpha1.SchemaInitStatements{
				Label:      b.Label,
				Statements: b.Statements,
			})
		}
	}

	switch s.connconfig.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		if options.GetTruncateBeforeInsert() {
			for k, v := range schemaTableMap {
				stmt, err := sqlmanager_mysql.BuildMysqlTruncateStatement(v.Schema, v.Table)
				if err != nil {
					return nil, err
				}
				truncateStmtsMap[k] = stmt
			}
		}

	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		if options.GetTruncateCascade() {
			for k, v := range schemaTableMap {
				stmt, err := sqlmanager_postgres.BuildPgTruncateCascadeStatement(v.Schema, v.Table)
				if err != nil {
					return nil, err
				}
				truncateStmtsMap[k] = stmt
			}
		} else if options.GetTruncateBeforeInsert() {
			return nil, nucleuserrors.NewNotImplemented("postgres truncate unsupported. table foreig keys required to build truncate statement.")
		}

	default:
		return nil, errors.New("unsupported connection config")
	}

	return &mgmtv1alpha1.GetConnectionInitStatementsResponse{
		TableInitStatements:     map[string]string{},
		TableTruncateStatements: truncateStmtsMap,
		SchemaInitStatements:    initSchemaStmts,
	}, nil
}

func (s *SQLConnectionDataService) GetTableConstraints(
	ctx context.Context,
) (*mgmtv1alpha1.GetConnectionTableConstraintsResponse, error) {
	schemaDbCols, err := s.GetSchema(ctx, nil)
	if err != nil {
		return nil, err
	}

	schemaMap := map[string]struct{}{}
	for _, s := range schemaDbCols {
		schemaMap[s.Schema] = struct{}{}
	}
	schemas := []string{}
	for s := range schemaMap {
		schemas = append(schemas, s)
	}

	db, err := s.sqlmanager.NewSqlConnection(ctx, connectionmanager.NewUniqueSession(), s.connection, s.logger)
	if err != nil {
		return nil, err
	}
	defer db.Db().Close()
	tableConstraints, err := db.Db().GetTableConstraintsBySchema(ctx, schemas)
	if err != nil {
		return nil, err
	}

	fkConstraintsMap := map[string]*mgmtv1alpha1.ForeignConstraintTables{}
	for tableName, d := range tableConstraints.ForeignKeyConstraints {
		fkConstraintsMap[tableName] = &mgmtv1alpha1.ForeignConstraintTables{
			Constraints: []*mgmtv1alpha1.ForeignConstraint{},
		}
		for _, constraint := range d {
			fkConstraintsMap[tableName].Constraints = append(fkConstraintsMap[tableName].Constraints, &mgmtv1alpha1.ForeignConstraint{
				Columns: constraint.Columns, NotNullable: constraint.NotNullable, ForeignKey: &mgmtv1alpha1.ForeignKey{
					Table:   constraint.ForeignKey.Table,
					Columns: constraint.ForeignKey.Columns,
				},
			})
		}
	}

	pkConstraintsMap := map[string]*mgmtv1alpha1.PrimaryConstraint{}
	for table, pks := range tableConstraints.PrimaryKeyConstraints {
		pkConstraintsMap[table] = &mgmtv1alpha1.PrimaryConstraint{
			Columns: pks,
		}
	}

	uniqueConstraintsMap := map[string]*mgmtv1alpha1.UniqueConstraints{}
	for table, uniqueConstraints := range tableConstraints.UniqueConstraints {
		uniqueConstraintsMap[table] = &mgmtv1alpha1.UniqueConstraints{
			Constraints: []*mgmtv1alpha1.UniqueConstraint{},
		}
		for _, uc := range uniqueConstraints {
			uniqueConstraintsMap[table].Constraints = append(uniqueConstraintsMap[table].Constraints, &mgmtv1alpha1.UniqueConstraint{
				Columns: uc,
			})
		}
	}

	uniqueIndexesMap := map[string]*mgmtv1alpha1.UniqueIndexes{}
	for table, uniqueIndexes := range tableConstraints.UniqueIndexes {
		uniqueIndexesMap[table] = &mgmtv1alpha1.UniqueIndexes{
			Indexes: []*mgmtv1alpha1.UniqueIndex{},
		}
		for _, ui := range uniqueIndexes {
			uniqueIndexesMap[table].Indexes = append(uniqueIndexesMap[table].Indexes, &mgmtv1alpha1.UniqueIndex{
				Columns: ui,
			})
		}
	}

	return &mgmtv1alpha1.GetConnectionTableConstraintsResponse{
		ForeignKeyConstraints: fkConstraintsMap,
		PrimaryKeyConstraints: pkConstraintsMap,
		UniqueConstraints:     uniqueConstraintsMap,
		UniqueIndexes:         uniqueIndexesMap,
	}, nil
}

func (s *SQLConnectionDataService) GetTableSchema(ctx context.Context, schema, table string) ([]*mgmtv1alpha1.DatabaseColumn, error) {
	db, err := s.sqlmanager.NewSqlConnection(ctx, connectionmanager.NewUniqueSession(), s.connection, s.logger)
	if err != nil {
		return nil, err
	}
	defer db.Db().Close()
	schematable := &sqlmanager_shared.SchemaTable{Schema: schema, Table: table}
	dbschema, err := db.Db().GetDatabaseTableSchemasBySchemasAndTables(ctx, []*sqlmanager_shared.SchemaTable{schematable})
	if err != nil {
		return nil, err
	}
	schemas := []*mgmtv1alpha1.DatabaseColumn{}
	for _, col := range dbschema {
		isNull := "NO"
		if col.IsNullable {
			isNull = "YES"
		}
		schemas = append(schemas, &mgmtv1alpha1.DatabaseColumn{
			Schema:     col.TableSchema,
			Table:      col.TableName,
			Column:     col.ColumnName,
			DataType:   col.DataType,
			IsNullable: isNull,
		})
	}
	return schemas, nil
}

func (s *SQLConnectionDataService) GetTableRowCount(ctx context.Context, schema, table string, whereClause *string) (int64, error) {
	db, err := s.sqlmanager.NewSqlConnection(ctx, connectionmanager.NewUniqueSession(), s.connection, s.logger)
	if err != nil {
		return 0, err
	}
	defer db.Db().Close()
	return db.Db().GetTableRowCount(ctx, schema, table, whereClause)
}

func (s *SQLConnectionDataService) areSchemaAndTableValid(ctx context.Context, schema, table string) error {
	schemas, err := s.GetTableSchema(ctx, schema, table)
	if err != nil {
		return err
	}

	if !isValidSchema(schema, schemas) || !isValidTable(table, schemas) {
		return nucleuserrors.NewBadRequest("must provide valid schema and table")
	}
	return nil
}

func isValidTable(table string, columns []*mgmtv1alpha1.DatabaseColumn) bool {
	for _, c := range columns {
		if c.Table == table {
			return true
		}
	}
	return false
}

func isValidSchema(schema string, columns []*mgmtv1alpha1.DatabaseColumn) bool {
	for _, c := range columns {
		if c.Schema == schema {
			return true
		}
	}
	return false
}
