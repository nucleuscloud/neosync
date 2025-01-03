package connectiondata

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	database_record_mapper "github.com/nucleuscloud/neosync/internal/database-record-mapper"
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

func (s *SQLConnectionDataService) areSchemaAndTableValid(ctx context.Context, schema, table string) error {
	schemas, err := s.GetSchema(ctx, nil)
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
