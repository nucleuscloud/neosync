package v1alpha1_connectionservice

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	jsonmodels "github.com/nucleuscloud/neosync/backend/internal/nucleusdb/json-models"
	"github.com/nucleuscloud/neosync/backend/internal/utils"
	k8s_utils "github.com/nucleuscloud/neosync/backend/internal/utils/k8s"
	neosyncdevv1alpha1 "github.com/nucleuscloud/neosync/k8s-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Service) CheckConnectionConfig(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.CheckConnectionConfigRequest],
) (*connect.Response[mgmtv1alpha1.CheckConnectionConfigResponse], error) {
	switch config := req.Msg.ConnectionConfig.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		var connectionString *string
		switch connectionConfig := config.PgConfig.ConnectionConfig.(type) {
		case *mgmtv1alpha1.PostgresConnectionConfig_Connection:
			connStr := nucleusdb.GetDbUrl(&nucleusdb.ConnectConfig{
				Host:     connectionConfig.Connection.Host,
				Port:     int(connectionConfig.Connection.Port),
				Database: connectionConfig.Connection.Name,
				User:     connectionConfig.Connection.User,
				Pass:     connectionConfig.Connection.Pass,
				SslMode:  connectionConfig.Connection.SslMode,
			})
			connectionString = &connStr
		case *mgmtv1alpha1.PostgresConnectionConfig_Url:
			connectionString = &connectionConfig.Url
		default:
			return nil, nucleuserrors.NewBadRequest("must provide valid postgres connection")
		}

		conn, err := pgx.Connect(ctx, *connectionString)
		if err != nil {
			msg := err.Error()
			return connect.NewResponse(&mgmtv1alpha1.CheckConnectionConfigResponse{
				IsConnected:     false,
				ConnectionError: &msg,
			}), nil
		}
		defer func() {
			if err := conn.Close(ctx); err != nil {
				log.Println("failed to close connection", err)
			}
		}()
		err = conn.Ping(ctx)
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
	default:
		return nil, nucleuserrors.NewNotImplemented("this connection config is not currently supported")
	}
}

func (s *Service) IsConnectionNameAvailable(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.IsConnectionNameAvailableRequest],
) (*connect.Response[mgmtv1alpha1.IsConnectionNameAvailableResponse], error) {
	accountId, err := s.verifyUserInAccount(ctx, req.Msg.AccountId)
	if err != nil {
		return nil, err
	}

	count, err := s.db.Q.IsConnectionNameAvailable(ctx, db_queries.IsConnectionNameAvailableParams{
		AccountId:      *accountId,
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
	// accountId, err := s.verifyUserInAccount(ctx, req.Msg.AccountId)
	// if err != nil {
	// 	return nil, err
	// }

	// connections, err := s.db.Q.GetConnectionsByAccount(ctx, *accountId)
	// if err != nil {
	// 	return nil, err
	// }

	// dtoConns := []*mgmtv1alpha1.Connection{}
	// for _, connection := range connections {
	// 	dtoConns = append(dtoConns, dtomaps.ToConnectionDto(&connection))
	// }

	return connect.NewResponse(&mgmtv1alpha1.GetConnectionsResponse{
		// Connections: dtoConns,
	}), nil
}

func (s *Service) GetConnection(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetConnectionRequest],
) (*connect.Response[mgmtv1alpha1.GetConnectionResponse], error) {
	conns := &neosyncdevv1alpha1.SqlConnectionList{}
	err := s.k8sclient.CustomResourceClient.List(ctx, conns, runtimeclient.InNamespace(s.k8sclient.Namespace), &runtimeclient.MatchingLabels{
		k8s_utils.NeosyncUuidLabel: req.Msg.Id,
	})
	if err != nil {
		return nil, err
	}
	if len(conns.Items) == 0 {
		return nil, nucleuserrors.NewNotFound(fmt.Sprintf("connection not found. id: %s", req.Msg.Id))
	}
	if len(conns.Items) > 1 {
		return nil, nucleuserrors.NewInternalError(fmt.Sprintf("more than 1 connection found. id: %s", req.Msg.Id))
	}

	conn := conns.Items[0]
	secretName := conn.Spec.Url.ValueFrom.SecretKeyRef.Name

	secrets, err := s.k8sclient.K8sClient.CoreV1().Secrets(s.k8sclient.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: ,

	})
	secrets, err := s.k8sclient.K8sClient.CoreV1().Secrets(s.k8sclient.Namespace).Get()
	if err != nil {
		return nil, err
	}

	filteredSecrets := utils.FilterSlice[corev1.Secret](secrets.Items, func(secret corev1.Secret) bool {
		return secret.Labels[k8s_utils.NeosyncUuidLabel] == req.Msg.Id
	})


	

	// TODO get connection first then get secret by reference namee
	//TODO verify ids match

	dto, err := dtomaps.ToConnectionDto(&conns.Items[0], &filteredSecrets[0])
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: dto,
	}), nil
}

// func (s *Service) CreateConnection(
// 	ctx context.Context,
// 	req *connect.Request[mgmtv1alpha1.CreateConnectionRequest],
// ) (*connect.Response[mgmtv1alpha1.CreateConnectionResponse], error) {
// 	accountUuid, err := s.verifyUserInAccount(ctx, req.Msg.AccountId)
// 	if err != nil {
// 		return nil, err
// 	}

// 	cc := &jsonmodels.ConnectionConfig{}
// 	if err := cc.FromDto(req.Msg.ConnectionConfig); err != nil {
// 		return nil, err
// 	}

// 	userUuid, err := s.getUserUuid(ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	connection, err := s.db.Q.CreateConnection(ctx, db_queries.CreateConnectionParams{
// 		AccountID:        *accountUuid,
// 		Name:             req.Msg.Name,
// 		ConnectionConfig: cc,
// 		CreatedByID:      *userUuid,
// 		UpdatedByID:      *userUuid,
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	return connect.NewResponse(&mgmtv1alpha1.CreateConnectionResponse{
// 		Connection: dtomaps.ToConnectionDto(&connection),
// 	}), nil
// }

func getPostgresConnectionUrl(c *mgmtv1alpha1.ConnectionConfig) (string, error) {
	switch config := c.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		var connectionString *string
		switch connectionConfig := config.PgConfig.ConnectionConfig.(type) {
		case *mgmtv1alpha1.PostgresConnectionConfig_Connection:
			connStr := nucleusdb.GetDbUrl(&nucleusdb.ConnectConfig{
				Host:     connectionConfig.Connection.Host,
				Port:     int(connectionConfig.Connection.Port),
				Database: connectionConfig.Connection.Name,
				User:     connectionConfig.Connection.User,
				Pass:     connectionConfig.Connection.Pass,
				SslMode:  connectionConfig.Connection.SslMode,
			})
			connectionString = &connStr
		case *mgmtv1alpha1.PostgresConnectionConfig_Url:
			connectionString = &connectionConfig.Url
		default:
			return "", nucleuserrors.NewBadRequest("must provide valid postgres connection")
		}
		return *connectionString, nil
	default:
		return "", nucleuserrors.NewNotImplemented("this connection config is not currently supported")
	}
}

func (s *Service) CreateConnection(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.CreateConnectionRequest],
) (*connect.Response[mgmtv1alpha1.CreateConnectionResponse], error) {
	// TODO handle failures or name collisions
	uuid := uuid.NewString()
	connectionString, err := getPostgresConnectionUrl(req.Msg.ConnectionConfig)
	if err != nil {
		return nil, err
	}

	sourceConnSecret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Msg.Name,
			Namespace: s.k8sclient.Namespace,
			Labels: map[string]string{
				k8s_utils.NeosyncUuidLabel: uuid,
			},
		},
		StringData: map[string]string{
			"url": connectionString,
		},
		Type: corev1.SecretTypeOpaque,
	}

	createdSecret, err := s.k8sclient.K8sClient.CoreV1().Secrets(s.k8sclient.Namespace).Create(ctx, sourceConnSecret, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	// make source connection
	sourceConnection := &neosyncdevv1alpha1.SqlConnection{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: s.k8sclient.Namespace,
			Name:      req.Msg.Name,
			Labels: map[string]string{
				k8s_utils.NeosyncUuidLabel: uuid,
			},
		},
		Spec: neosyncdevv1alpha1.SqlConnectionSpec{
			Driver: neosyncdevv1alpha1.PostgresDriver,
			Url: neosyncdevv1alpha1.SqlConnectionUrl{
				ValueFrom: &neosyncdevv1alpha1.SqlConnectionUrlSource{
					SecretKeyRef: &neosyncdevv1alpha1.ConfigSelector{
						Name: createdSecret.Name,
						Key:  "url",
					},
				},
			},
		},
	}
	err = s.k8sclient.CustomResourceClient.Create(ctx, sourceConnection)
	if err != nil {
		return nil, err
	}

	dto, err := dtomaps.ToConnectionDto(sourceConnection, sourceConnSecret)
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
	connection, err := s.db.Q.GetConnectionById(ctx, connectionUuid)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return nil, nucleuserrors.NewNotFound("unable to find connection by id")
	}

	_, err = s.verifyUserInAccount(ctx, nucleusdb.UUIDString(connection.AccountID))
	if err != nil {
		return nil, err
	}

	cc := &jsonmodels.ConnectionConfig{}
	if err := cc.FromDto(req.Msg.ConnectionConfig); err != nil {
		return nil, err
	}

	userUuid, err := s.getUserUuid(ctx)
	if err != nil {
		return nil, err
	}

	connection, err = s.db.Q.UpdateConnection(ctx, db_queries.UpdateConnectionParams{
		ID:               connection.ID,
		ConnectionConfig: cc,
		UpdatedByID:      *userUuid,
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.UpdateConnectionResponse{
		// Connection: dtomaps.ToConnectionDto(&connection),
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

	connection, err := s.db.Q.GetConnectionById(ctx, idUuid)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return connect.NewResponse(&mgmtv1alpha1.DeleteConnectionResponse{}), nil
	}

	_, err = s.verifyUserInAccount(ctx, nucleusdb.UUIDString(connection.AccountID))
	if err != nil {
		return nil, err
	}

	err = s.db.Q.RemoveConnectionById(ctx, connection.ID)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mgmtv1alpha1.DeleteConnectionResponse{}), nil
}

func (s *Service) verifyUserInAccount(
	ctx context.Context,
	accountId string,
) (*pgtype.UUID, error) {
	resp, err := s.userAccountService.IsUserInAccount(ctx, connect.NewRequest(&mgmtv1alpha1.IsUserInAccountRequest{AccountId: accountId}))
	if err != nil {
		return nil, err
	}
	if !resp.Msg.Ok {
		return nil, nucleuserrors.NewForbidden("user in not in requested account")
	}

	accountUuid, err := nucleusdb.ToUuid(accountId)
	if err != nil {
		return nil, err
	}
	return &accountUuid, nil
}

func (s *Service) getUserUuid(
	ctx context.Context,
) (*pgtype.UUID, error) {
	user, err := s.userAccountService.GetUser(ctx, connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	if err != nil {
		return nil, err
	}
	userUuid, err := nucleusdb.ToUuid(user.Msg.UserId)
	if err != nil {
		return nil, err
	}
	return &userUuid, nil
}
