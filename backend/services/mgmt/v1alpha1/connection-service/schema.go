package v1alpha1_connectionservice

import (
	"context"
	"database/sql"
	"fmt"

	"connectrpc.com/connect"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
)

type DatabaseSchema struct {
	TableSchema string `db:"table_schema,omitempty"`
	TableName   string `db:"table_name,omitempty"`
	ColumnName  string `db:"column_name,omitempty"`
	DataType    string `db:"data_type,omitempty"`
}

func (s *Service) GetConnectionSchema(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetConnectionSchemaRequest],
) (*connect.Response[mgmtv1alpha1.GetConnectionSchemaResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("connectionId", req.Msg.Id)
	connection, err := s.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
		Id: req.Msg.Id,
	}))
	if err != nil {
		return nil, err
	}

	connCfg := connection.Msg.Connection.ConnectionConfig
	connectionString, err := s.getConnectionUrl(connCfg)
	if err != nil {
		return nil, err
	}

	switch connCfg.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		conn, err := pgx.Connect(ctx, connectionString)
		if err != nil {
			logger.Error("unable to connect", err)
			return nil, err
		}
		defer func() {
			if err := conn.Close(ctx); err != nil {
				logger.Error(fmt.Errorf("failed to close postgres connection: %w", err).Error())
			}
		}()

		dbSchema, err := getPostgresDatabaseSchema(ctx, conn)
		if err != nil {
			return nil, err
		}

		return connect.NewResponse(&mgmtv1alpha1.GetConnectionSchemaResponse{
			Schemas: ToDatabaseColumn(dbSchema),
		}), nil

	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		conn, err := sql.Open("mysql", connectionString)
		if err != nil {
			logger.Error("unable to connect", err)
			return nil, err
		}
		defer func() {
			if err := conn.Close(); err != nil {
				logger.Error(fmt.Errorf("failed to close mysql connection: %w", err).Error())
			}
		}()

		dbSchema, err := getMysqlDatabaseSchema(ctx, conn)
		if err != nil {
			return nil, err
		}

		return connect.NewResponse(&mgmtv1alpha1.GetConnectionSchemaResponse{
			Schemas: ToDatabaseColumn(dbSchema),
		}), nil

	default:
		return nil, nucleuserrors.NewNotImplemented("this connection config is not currently supported")
	}
}

const (
	getPostgresTableSchemaSql = `-- name: GetPostgresTableSchema
	SELECT
	c.table_schema,
	c.table_name,
	c.column_name,
	c.data_type
	FROM
		information_schema.columns AS c
		JOIN information_schema.tables AS t ON c.table_schema = t.table_schema
			AND c.table_name = t.table_name
	WHERE
		c.table_schema NOT IN ('pg_catalog', 'information_schema')
		AND t.table_type = 'BASE TABLE';
`

	getMysqlTableSchemaSql = `-- name: GetMysqlTableSchema
	SELECT
	c.table_schema,
	c.table_name,
	c.column_name,
	c.data_type
	FROM
		information_schema.columns AS c
		JOIN information_schema.tables AS t ON c.table_schema = t.table_schema
			AND c.table_name = t.table_name
	WHERE
		t.table_type = 'BASE TABLE';
`
)

func getPostgresDatabaseSchema(ctx context.Context, conn *pgx.Conn) ([]DatabaseSchema, error) {
	rows, err := conn.Query(ctx, getPostgresTableSchemaSql)
	if err != nil && !isNoRows(err) {
		return nil, err
	}
	if err != nil && isNoRows(err) {
		return []DatabaseSchema{}, nil
	}

	output := []DatabaseSchema{}
	for rows.Next() {
		var o DatabaseSchema
		err := rows.Scan(
			&o.TableSchema,
			&o.TableName,
			&o.ColumnName,
			&o.DataType,
		)
		if err != nil {
			return nil, err
		}
		output = append(output, o)
	}
	return output, nil
}

func getMysqlDatabaseSchema(ctx context.Context, conn *sql.DB) ([]DatabaseSchema, error) {
	rows, err := conn.QueryContext(ctx, getMysqlTableSchemaSql)
	if err != nil && !isNoRows(err) {
		return nil, err
	}
	if err != nil && isNoRows(err) {
		return []DatabaseSchema{}, nil
	}

	output := []DatabaseSchema{}
	for rows.Next() {
		var o DatabaseSchema
		err := rows.Scan(
			&o.TableSchema,
			&o.TableName,
			&o.ColumnName,
			&o.DataType,
		)
		if err != nil {
			return nil, err
		}
		output = append(output, o)
	}
	return output, nil
}

func isNoRows(err error) bool {
	return err != nil && err == sql.ErrNoRows || err == pgx.ErrNoRows
}

func ToDatabaseColumn(
	input []DatabaseSchema,
) []*mgmtv1alpha1.DatabaseColumn {
	columns := []*mgmtv1alpha1.DatabaseColumn{}
	for _, col := range input {
		columns = append(columns, &mgmtv1alpha1.DatabaseColumn{
			Schema:   col.TableSchema,
			Table:    col.TableName,
			Column:   col.ColumnName,
			DataType: col.DataType,
		})
	}
	return columns
}
