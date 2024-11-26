package v1alpha1_connectionservice

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"connectrpc.com/connect"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
	dbconnectconfig "github.com/nucleuscloud/neosync/backend/pkg/dbconnect-config"
	pg_models "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"
	"golang.org/x/sync/errgroup"

	"go.mongodb.org/mongo-driver/bson"
)

type connInput struct {
	cc *mgmtv1alpha1.ConnectionConfig
	id string
}

func (c *connInput) GetId() string {
	return c.id
}
func (c *connInput) GetConnectionConfig() *mgmtv1alpha1.ConnectionConfig {
	return c.cc
}

func (s *Service) CheckConnectionConfig(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.CheckConnectionConfigRequest],
) (*connect.Response[mgmtv1alpha1.CheckConnectionConfigResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)

	switch req.Msg.GetConnectionConfig().GetConfig().(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig, *mgmtv1alpha1.ConnectionConfig_MysqlConfig, *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
		role, err := getDbRoleFromConnectionConfig(req.Msg.GetConnectionConfig(), logger)
		if err != nil {
			return nil, err
		}
		s.sqlmanager.NewSqlConnection(ctx, &mgmtv1alpha1.Connection{ConnectionConfig: req.Msg.GetConnectionConfig()}, logger)
		db, err := s.sqlmanager.NewSqlConnection(ctx, &connInput{cc: req.Msg.GetConnectionConfig(), id: uuid.NewString()}, logger)
		if err != nil {
			return nil, err
		}
		defer db.Db().Close()
		schematablePrivsMap, err := db.Db().GetRolePermissionsMap(ctx)
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
		db, err := s.mongoconnector.NewFromConnectionConfig(req.Msg.GetConnectionConfig(), logger)
		if err != nil {
			return nil, err
		}
		client, err := db.Open(ctx)
		if err != nil {
			return nil, err
		}
		defer func() {
			err := db.Close(ctx)
			if err != nil {
				logger.Warn(fmt.Sprintf("unable to close all mongodb connections: %s", err.Error()))
			}
		}()

		dbnames, err := client.ListDatabaseNames(ctx, bson.D{})
		if err != nil {
			errmsg := err.Error()
			return connect.NewResponse(&mgmtv1alpha1.CheckConnectionConfigResponse{
				IsConnected:     false,
				ConnectionError: &errmsg,
				Privileges:      []*mgmtv1alpha1.ConnectionRolePrivilege{},
			}), nil
		}

		collectionsMap := map[string][]string{}
		collectionMu := sync.Mutex{}

		errgrp, errctx := errgroup.WithContext(ctx)
		for _, dbname := range dbnames {
			dbname := dbname
			errgrp.Go(func() error {
				collnames, err := client.Database(dbname).ListCollectionNames(errctx, bson.D{})
				if err != nil {
					return fmt.Errorf("unable to retrieve collection names for database %q: %w", dbname, err)
				}
				collectionMu.Lock()
				defer collectionMu.Unlock()
				collectionsMap[dbname] = append(collectionsMap[dbname], collnames...)
				return nil
			})
		}
		err = errgrp.Wait()
		if err != nil {
			return nil, err
		}

		privs := []*mgmtv1alpha1.ConnectionRolePrivilege{}
		for dbname, collections := range collectionsMap {
			for _, collection := range collections {
				privs = append(privs, &mgmtv1alpha1.ConnectionRolePrivilege{
					Schema:        dbname,
					Table:         collection,
					Grantee:       "",
					PrivilegeType: []string{},
				})
			}
		}

		return connect.NewResponse(&mgmtv1alpha1.CheckConnectionConfigResponse{
			IsConnected:     true,
			ConnectionError: nil,
			Privileges:      privs,
		}), nil
	case *mgmtv1alpha1.ConnectionConfig_DynamodbConfig:
		client, err := s.awsManager.NewDynamoDbClient(ctx, req.Msg.GetConnectionConfig().GetDynamodbConfig())
		if err != nil {
			return nil, err
		}
		tableNames, err := client.ListAllTables(ctx, &dynamodb.ListTablesInput{})
		if err != nil {
			return nil, err
		}

		privs := []*mgmtv1alpha1.ConnectionRolePrivilege{}
		for _, tableName := range tableNames {
			privs = append(privs, &mgmtv1alpha1.ConnectionRolePrivilege{
				Schema:        "",
				Table:         tableName,
				Grantee:       "",
				PrivilegeType: []string{},
			})
		}

		return connect.NewResponse(&mgmtv1alpha1.CheckConnectionConfigResponse{
			IsConnected:     true,
			ConnectionError: nil,
			Privileges:      privs,
		}), nil
	default:
		return nil, nucleuserrors.NewBadRequest(fmt.Errorf("this method does not support this connection type %T: %w", req.Msg.GetConnectionConfig().GetConfig(), errors.ErrUnsupported).Error())
	}
}

func (s *Service) CheckConnectionConfigById(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.CheckConnectionConfigByIdRequest],
) (*connect.Response[mgmtv1alpha1.CheckConnectionConfigByIdResponse], error) {
	connResp, err := s.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
		Id: req.Msg.Id,
	}))
	if err != nil {
		return nil, err
	}

	resp, err := s.CheckConnectionConfig(ctx, connect.NewRequest(&mgmtv1alpha1.CheckConnectionConfigRequest{
		ConnectionConfig: connResp.Msg.GetConnection().ConnectionConfig,
	}))
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.CheckConnectionConfigByIdResponse{
		IsConnected:     resp.Msg.GetIsConnected(),
		ConnectionError: resp.Msg.ConnectionError,
		Privileges:      resp.Msg.GetPrivileges(),
	}), nil
}

func getDbRoleFromConnectionConfig(cconfig *mgmtv1alpha1.ConnectionConfig, logger *slog.Logger) (string, error) {
	if cconfig == nil {
		return "", errors.New("connection config was nil, unable to retrieve db role")
	}

	switch typedconfig := cconfig.GetConfig().(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		parsedCfg, err := dbconnectconfig.NewFromPostgresConnection(typedconfig, nil, logger)
		if err != nil {
			return "", fmt.Errorf("unable to parse pg connection: %w", err)
		}
		return parsedCfg.GetUser(), nil
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		parsedCfg, err := dbconnectconfig.NewFromMysqlConnection(typedconfig, nil, logger, false)
		if err != nil {
			return "", fmt.Errorf("unable to parse mysql connection: %w", err)
		}
		return parsedCfg.GetUser(), nil
	case *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
		parsedCfg, err := dbconnectconfig.NewFromMssqlConnection(typedconfig, nil)
		if err != nil {
			return "", fmt.Errorf("unable to parse mssql connection: %w", err)
		}
		return parsedCfg.GetUser(), nil
	default:
		return "", fmt.Errorf("invalid database connection config (%T) for retrieving db role: %w", typedconfig, errors.ErrUnsupported)
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
	idUuid, err := neosyncdb.ToUuid(req.Msg.Id)
	if err != nil {
		return nil, err
	}

	connection, err := s.db.Q.GetConnectionById(ctx, s.db.Db, idUuid)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return nil, nucleuserrors.NewNotFound("unable to find connection by id")
	}

	_, err = s.verifyUserInAccount(ctx, neosyncdb.UUIDString(connection.AccountID))
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
	connectionUuid, err := neosyncdb.ToUuid(req.Msg.Id)
	if err != nil {
		return nil, err
	}
	connection, err := s.db.Q.GetConnectionById(ctx, s.db.Db, connectionUuid)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return nil, nucleuserrors.NewNotFound("unable to find connection by id")
	}

	_, err = s.verifyUserInAccount(ctx, neosyncdb.UUIDString(connection.AccountID))
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
	idUuid, err := neosyncdb.ToUuid(req.Msg.Id)
	if err != nil {
		return nil, err
	}

	connection, err := s.db.Q.GetConnectionById(ctx, s.db.Db, idUuid)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return connect.NewResponse(&mgmtv1alpha1.DeleteConnectionResponse{}), nil
	}

	_, err = s.verifyUserInAccount(ctx, neosyncdb.UUIDString(connection.AccountID))
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
	defer neosyncdb.HandleSqlRollback(tx, logger)

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
