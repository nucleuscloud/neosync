package v1alpha1_connectionservice

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"connectrpc.com/connect"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	pg_models "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"
)

const (
	mysqlDriver    = "mysql"
	postgresDriver = "postgres"
)

func (s *Service) CheckConnectionConfig(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.CheckConnectionConfigRequest],
) (*connect.Response[mgmtv1alpha1.CheckConnectionConfigResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	connectionTimeout := uint32(5)
	conn, err := s.sqlConnector.NewDbFromConnectionConfig(req.Msg.ConnectionConfig, &connectionTimeout, logger)
	if err != nil {
		return nil, err
	}
	db, err := conn.Open()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := conn.Close(); err != nil {
			logger.Error(fmt.Errorf("failed to close database connection: %w", err).Error())
		}
	}()

	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err = db.PingContext(cctx)
	if err != nil {
		msg := err.Error()
		return connect.NewResponse(&mgmtv1alpha1.CheckConnectionConfigResponse{
			IsConnected:     false,
			ConnectionError: &msg,
		}), nil
	}
	return connect.NewResponse(&mgmtv1alpha1.CheckConnectionConfigResponse{
		IsConnected:     true,
		ConnectionError: nil,
	}), nil
}

func (s *Service) IsConnectionNameAvailable(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.IsConnectionNameAvailableRequest],
) (*connect.Response[mgmtv1alpha1.IsConnectionNameAvailableResponse], error) {
	accountUuid, err := s.verifyUserInAccount(ctx, req.Msg.AccountId)
	if err != nil {
		return nil, err
	}

	count, err := s.db.Q.IsConnectionNameAvailable(ctx, s.db.Db, db_queries.IsConnectionNameAvailableParams{
		AccountId:      *accountUuid,
		ConnectionName: req.Msg.ConnectionName,
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.IsConnectionNameAvailableResponse{
		IsAvailable: count == 0,
	}), nil
}

func (s *Service) GetConnections(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetConnectionsRequest],
) (*connect.Response[mgmtv1alpha1.GetConnectionsResponse], error) {
	accountUuid, err := s.verifyUserInAccount(ctx, req.Msg.AccountId)
	if err != nil {
		return nil, err
	}

	connections, err := s.db.Q.GetConnectionsByAccount(ctx, s.db.Db, *accountUuid)
	if err != nil {
		return nil, err
	}

	dtoConns := []*mgmtv1alpha1.Connection{}
	for idx := range connections {
		connection := connections[idx]
		dtoConns = append(dtoConns, dtomaps.ToConnectionDto(&connection))
	}

	return connect.NewResponse(&mgmtv1alpha1.GetConnectionsResponse{
		Connections: dtoConns,
	}), nil
}

func (s *Service) GetConnection(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetConnectionRequest],
) (*connect.Response[mgmtv1alpha1.GetConnectionResponse], error) {
	idUuid, err := nucleusdb.ToUuid(req.Msg.Id)
	if err != nil {
		return nil, err
	}

	connection, err := s.db.Q.GetConnectionById(ctx, s.db.Db, idUuid)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return nil, nucleuserrors.NewNotFound("unable to find connection by id")
	}

	_, err = s.verifyUserInAccount(ctx, nucleusdb.UUIDString(connection.AccountID))
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: dtomaps.ToConnectionDto(&connection),
	}), nil
}

func (s *Service) CreateConnection(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.CreateConnectionRequest],
) (*connect.Response[mgmtv1alpha1.CreateConnectionResponse], error) {
	cc := &pg_models.ConnectionConfig{}
	if err := cc.FromDto(req.Msg.ConnectionConfig); err != nil {
		return nil, err
	}

	accountUuid, err := s.verifyUserInAccount(ctx, req.Msg.AccountId)
	if err != nil {
		return nil, err
	}

	userUuid, err := s.getUserUuid(ctx)
	if err != nil {
		return nil, err
	}

	connection, err := s.db.Q.CreateConnection(ctx, s.db.Db, db_queries.CreateConnectionParams{
		AccountID:        *accountUuid,
		Name:             req.Msg.Name,
		ConnectionConfig: cc,
		CreatedByID:      *userUuid,
		UpdatedByID:      *userUuid,
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.CreateConnectionResponse{
		Connection: dtomaps.ToConnectionDto(&connection),
	}), nil
}

func (s *Service) UpdateConnection(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.UpdateConnectionRequest],
) (*connect.Response[mgmtv1alpha1.UpdateConnectionResponse], error) {
	connectionUuid, err := nucleusdb.ToUuid(req.Msg.Id)
	if err != nil {
		return nil, err
	}
	connection, err := s.db.Q.GetConnectionById(ctx, s.db.Db, connectionUuid)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return nil, nucleuserrors.NewNotFound("unable to find connection by id")
	}

	_, err = s.verifyUserInAccount(ctx, nucleusdb.UUIDString(connection.AccountID))
	if err != nil {
		return nil, err
	}

	userUuid, err := s.getUserUuid(ctx)
	if err != nil {
		return nil, err
	}

	cc := &pg_models.ConnectionConfig{}
	if err := cc.FromDto(req.Msg.ConnectionConfig); err != nil {
		return nil, err
	}

	connection, err = s.db.Q.UpdateConnection(ctx, s.db.Db, db_queries.UpdateConnectionParams{
		ID:               connection.ID,
		ConnectionConfig: cc,
		UpdatedByID:      *userUuid,
		Name:             req.Msg.Name,
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.UpdateConnectionResponse{
		Connection: dtomaps.ToConnectionDto(&connection),
	}), nil
}

func (s *Service) DeleteConnection(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.DeleteConnectionRequest],
) (*connect.Response[mgmtv1alpha1.DeleteConnectionResponse], error) {
	idUuid, err := nucleusdb.ToUuid(req.Msg.Id)
	if err != nil {
		return nil, err
	}

	connection, err := s.db.Q.GetConnectionById(ctx, s.db.Db, idUuid)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return connect.NewResponse(&mgmtv1alpha1.DeleteConnectionResponse{}), nil
	}

	_, err = s.verifyUserInAccount(ctx, nucleusdb.UUIDString(connection.AccountID))
	if err != nil {
		return nil, err
	}

	err = s.db.Q.RemoveConnectionById(ctx, s.db.Db, connection.ID)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mgmtv1alpha1.DeleteConnectionResponse{}), nil
}

func (s *Service) CheckSqlQuery(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.CheckSqlQueryRequest],
) (*connect.Response[mgmtv1alpha1.CheckSqlQueryResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("connectionId", req.Msg.Id)
	connection, err := s.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{Id: req.Msg.Id}))
	if err != nil {
		return nil, err
	}

	conn, err := s.sqlConnector.NewDbFromConnectionConfig(connection.Msg.Connection.ConnectionConfig, nil, logger)
	if err != nil {
		return nil, err
	}

	db, err := conn.Open()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := conn.Close(); err != nil {
			logger.Error(fmt.Errorf("failed to close connection: %w", err).Error())
		}
	}()
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer nucleusdb.HandleSqlRollback(tx, logger)

	_, err = tx.PrepareContext(ctx, req.Msg.Query)
	var errorMsg *string
	if err != nil {
		msg := err.Error()
		errorMsg = &msg
	}
	return connect.NewResponse(&mgmtv1alpha1.CheckSqlQueryResponse{
		IsValid:      err == nil,
		ErorrMessage: errorMsg,
	}), nil
}
