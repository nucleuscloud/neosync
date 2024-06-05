package v1alpha1_connectionservice

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"connectrpc.com/connect"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	pg_models "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"
)

func (s *Service) CheckConnectionConfig(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.CheckConnectionConfigRequest],
) (*connect.Response[mgmtv1alpha1.CheckConnectionConfigResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)

	switch req.Msg.GetConnectionConfig().GetConfig().(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig, *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		role, err := getDbRoleFromConnectionConfig(req.Msg.GetConnectionConfig())
		if err != nil {
			return nil, err
		}
		connTimeout := 5
		db, err := s.sqlmanager.NewSqlDbFromConnectionConfig(ctx, logger, req.Msg.GetConnectionConfig(), &connTimeout)
		if err != nil {
			return nil, err
		}
		defer db.Db.Close()
		schematablePrivsMap, err := db.Db.GetRolePermissionsMap(ctx, role)
		if err != nil {
			errmsg := err.Error()
			return connect.NewResponse(&mgmtv1alpha1.CheckConnectionConfigResponse{
				IsConnected:     false,
				ConnectionError: &errmsg,
				Privileges:      nil,
			}), nil
		}

		privs := []*mgmtv1alpha1.ConnectionRolePrivilege{}
		for key, permissions := range schematablePrivsMap {
			parts := strings.SplitN(key, ".", 2)
			schema, table := parts[0], parts[1]
			privs = append(privs, &mgmtv1alpha1.ConnectionRolePrivilege{
				Grantee:       role,
				Schema:        schema,
				Table:         table,
				PrivilegeType: permissions,
			})
		}
		return connect.NewResponse(&mgmtv1alpha1.CheckConnectionConfigResponse{
			IsConnected:     true,
			ConnectionError: nil,
			Privileges:      privs,
		}), nil

	case *mgmtv1alpha1.ConnectionConfig_MongoConfig:
		return connect.NewResponse(&mgmtv1alpha1.CheckConnectionConfigResponse{
			IsConnected:     false,
			ConnectionError: nil,
			Privileges:      []*mgmtv1alpha1.ConnectionRolePrivilege{},
		}), nil
	default:
		return nil, fmt.Errorf("this method does not support this connection type %T: %w", req.Msg.GetConnectionConfig().GetConfig(), errors.ErrUnsupported)
	}
}

func getDbRoleFromConnectionConfig(cconfig *mgmtv1alpha1.ConnectionConfig) (string, error) {
	if cconfig == nil {
		return "", errors.New("connection config was nil, unable to retrieve db role")
	}

	switch typedconfig := cconfig.GetConfig().(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		return getPostgresUserFromConnectionConfig(typedconfig.PgConfig)
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		return getMysqlUserFromConnectionConfig(typedconfig.MysqlConfig)
	}
	return "", fmt.Errorf("invalid database connection config for retrieving db role: %w", errors.ErrUnsupported)
}

func getPostgresUserFromConnectionConfig(pgconfig *mgmtv1alpha1.PostgresConnectionConfig) (string, error) {
	switch config := pgconfig.ConnectionConfig.(type) {
	case *mgmtv1alpha1.PostgresConnectionConfig_Connection:
		return config.Connection.User, nil
	case *mgmtv1alpha1.PostgresConnectionConfig_Url:
		u, err := url.Parse(config.Url)
		if err != nil {
			var urlErr *url.Error
			if errors.As(err, &urlErr) {
				return "", fmt.Errorf("unable to parse postgres url [%s]: %w", urlErr.Op, urlErr.Err)
			}
			return "", fmt.Errorf("unable to parse postgres url: %w", err)
		}
		return u.User.Username(), nil
	default:
		return "", fmt.Errorf("unable to parse connection url from postgres config: %T", config)
	}
}

func getMysqlUserFromConnectionConfig(pgconfig *mgmtv1alpha1.MysqlConnectionConfig) (string, error) {
	switch config := pgconfig.ConnectionConfig.(type) {
	case *mgmtv1alpha1.MysqlConnectionConfig_Connection:
		return config.Connection.User, nil
	case *mgmtv1alpha1.MysqlConnectionConfig_Url:
		u, err := url.Parse(config.Url)
		if err != nil {
			var urlErr *url.Error
			if errors.As(err, &urlErr) {
				return "", fmt.Errorf("unable to parse postgres url [%s]: %w", urlErr.Op, urlErr.Err)
			}
			return "", fmt.Errorf("unable to parse postgres url: %w", err)
		}
		return u.User.Username(), nil
	default:
		return "", fmt.Errorf("unable to parse connection url from postgres config: %T", config)
	}
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
		dto, err := dtomaps.ToConnectionDto(&connection)
		if err != nil {
			return nil, err
		}
		dtoConns = append(dtoConns, dto)
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
	dto, err := dtomaps.ToConnectionDto(&connection)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: dto,
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
	dto, err := dtomaps.ToConnectionDto(&connection)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mgmtv1alpha1.CreateConnectionResponse{
		Connection: dto,
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
	dto, err := dtomaps.ToConnectionDto(&connection)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mgmtv1alpha1.UpdateConnectionResponse{
		Connection: dto,
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
