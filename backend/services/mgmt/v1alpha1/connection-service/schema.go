package v1alpha1_connectionservice

import (
	"context"
	"database/sql"
	"fmt"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
)

type DatabaseSchema struct {
	TableSchema string `db:"table_schema,omitempty"`
	TableName   string `db:"table_name,omitempty"`
	ColumnName  string `db:"column_name,omitempty"`
	DataType    string `db:"data_type,omitempty"`
}

type DatabaseTableConstraints struct {
	Name       string `db:"name,omitempty"`
	Type       string `db:"contype,omitempty"`
	Definition string `db:"definition,omitempty"`
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
	connectionString, err := s.GetConnectionUrl(connCfg)
	if err != nil {
		return nil, err
	}

	conn, err := pgx.Connect(ctx, connectionString)
	if err != nil {
		logger.Error("unable to connect", err)
		return nil, err
	}
	defer func() {
		if err := conn.Close(ctx); err != nil {
			logger.Error(fmt.Errorf("failed to close connection: %w", err).Error())
		}
	}()

	dbSchema, err := getDatabaseSchema(ctx, conn)
	if err != nil {
		return nil, err
	}

	schema := []*mgmtv1alpha1.DatabaseColumn{}
	for _, row := range dbSchema {
		schema = append(schema, &mgmtv1alpha1.DatabaseColumn{
			Schema:   row.TableSchema,
			Table:    row.TableName,
			Column:   row.ColumnName,
			DataType: row.DataType,
		})
	}

	return connect.NewResponse(&mgmtv1alpha1.GetConnectionSchemaResponse{
		Schemas: schema,
	}), nil
}

func getDatabaseSchema(ctx context.Context, conn *pgx.Conn) ([]DatabaseSchema, error) {
	rows, err := conn.Query(ctx, `
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
			c.table_schema NOT IN('pg_catalog', 'information_schema')
			AND t.table_type = 'BASE TABLE';
	`)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err != nil && err == sql.ErrNoRows {
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
