package v1alpha1_connectionservice

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	neosync_k8sclient "github.com/nucleuscloud/neosync/backend/internal/k8s/client"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	conn_utils "github.com/nucleuscloud/neosync/backend/internal/utils/connections"
	k8s_utils "github.com/nucleuscloud/neosync/backend/internal/utils/k8s"
	neosyncdevv1alpha1 "github.com/nucleuscloud/neosync/k8s-operator/api/v1alpha1"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
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
	connection := &neosyncdevv1alpha1.SqlConnection{}
	err := s.k8sclient.CustomResourceClient.Get(
		ctx,
		types.NamespacedName{Name: req.Msg.ConnectionName, Namespace: s.cfg.JobConfigNamespace},
		connection,
	)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	} else if err != nil && errors.IsNotFound(err) {
		return connect.NewResponse(&mgmtv1alpha1.IsConnectionNameAvailableResponse{
			IsAvailable: true,
		}), nil
	}

	return connect.NewResponse(&mgmtv1alpha1.IsConnectionNameAvailableResponse{
		IsAvailable: false,
	}), nil
}

func (s *Service) GetConnections(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetConnectionsRequest],
) (*connect.Response[mgmtv1alpha1.GetConnectionsResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	conns := &neosyncdevv1alpha1.SqlConnectionList{}
	err := s.k8sclient.CustomResourceClient.List(ctx, conns, runtimeclient.InNamespace(s.cfg.JobConfigNamespace))
	if err != nil && !errors.IsNotFound(err) {
		logger.Error("unable to retrieve connections")
		return nil, err
	} else if err != nil && errors.IsNotFound(err) {
		return connect.NewResponse(&mgmtv1alpha1.GetConnectionsResponse{
			Connections: []*mgmtv1alpha1.Connection{},
		}), nil
	}
	if len(conns.Items) == 0 {
		return connect.NewResponse(&mgmtv1alpha1.GetConnectionsResponse{
			Connections: []*mgmtv1alpha1.Connection{},
		}), nil
	}

	secrets, err := s.k8sclient.K8sClient.CoreV1().Secrets(s.cfg.JobConfigNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: k8s_utils.NeosyncUuidLabel,
	})
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	secretsMap := map[string]*corev1.Secret{}
	for i := range secrets.Items {
		secret := secrets.Items[i]
		secretsMap[secret.Name] = &secret
	}

	dtoConns := []*mgmtv1alpha1.Connection{}
	for i := range conns.Items {
		conn := conns.Items[i]
		connId := conn.Labels[k8s_utils.NeosyncUuidLabel]
		var secret *corev1.Secret
		if conn.Spec.Url.ValueFrom != nil {
			secretName := conn.Spec.Url.ValueFrom.SecretKeyRef.Name
			secretEntry, ok := secretsMap[secretName]
			if ok {
				secretId := secretEntry.Labels[k8s_utils.NeosyncUuidLabel]
				if connId != secretId {
					msg := fmt.Sprintf("connection and secret uuid mismatch. connId: %s secretId: %s", connId, secretId)
					return nil, nucleuserrors.NewInternalError(msg)
				}
				secret = secretEntry
			}
		}
		dto, err := dtomaps.ToConnectionDto(&conn, secret)
		if err != nil {
			return nil, err
		}
		dtoConns = append(dtoConns, dto)
	}

	return connect.NewResponse(&mgmtv1alpha1.GetConnectionsResponse{
		Connections: dtoConns,
	}), nil
}

func (s *Service) GetConnectionsByNames(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetConnectionsByNamesRequest],
) (*connect.Response[mgmtv1alpha1.GetConnectionsByNamesResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	connsChan := make(chan *neosyncdevv1alpha1.SqlConnection, len(req.Msg.Names))
	errs, errCtx := errgroup.WithContext(ctx)
	for _, name := range req.Msg.Names {
		name := name
		errs.Go(func() error {
			conn := &neosyncdevv1alpha1.SqlConnection{}
			err := s.k8sclient.CustomResourceClient.Get(errCtx, types.NamespacedName{Name: name, Namespace: s.cfg.JobConfigNamespace}, conn)
			if err != nil && !errors.IsNotFound(err) {
				return err
			} else if err != nil && errors.IsNotFound(err) {
				logger.Warn("connection not found", "connectionName", name)
				return nil
			}
			select {
			case connsChan <- conn:
				return nil
			case <-errCtx.Done():
				return errCtx.Err()
			}
		})
	}
	err := errs.Wait()
	if err != nil {
		return nil, err
	}
	close(connsChan)

	conns := []*neosyncdevv1alpha1.SqlConnection{}
	secretNames := []string{}
	for conn := range connsChan {
		if conn.Spec.Url.ValueFrom != nil {
			secretNames = append(secretNames, conn.Spec.Url.ValueFrom.SecretKeyRef.Name)
		}
		conns = append(conns, conn)
	}

	secretsChan := make(chan *corev1.Secret, len(secretNames))
	errs, errCtx = errgroup.WithContext(ctx)
	for _, name := range secretNames {
		name := name
		errs.Go(func() error {
			secret, err := getConnectionSecretByName(errCtx, logger, s.k8sclient, name, s.cfg.JobConfigNamespace)
			if err != nil && !errors.IsNotFound(err) {
				return err
			} else if err != nil && errors.IsNotFound(err) {
				return nil
			}
			select {
			case secretsChan <- secret:
				return nil
			case <-errCtx.Done():
				return errCtx.Err()
			}
		})
	}
	err = errs.Wait()
	if err != nil {
		return nil, err
	}
	close(secretsChan)

	secretsMap := map[string]*corev1.Secret{}
	for secret := range secretsChan {
		secretsMap[secret.Name] = secret
	}

	dtoConns := []*mgmtv1alpha1.Connection{}
	for _, conn := range conns {
		connId := conn.Labels[k8s_utils.NeosyncUuidLabel]
		var secret *corev1.Secret
		if conn.Spec.Url.ValueFrom != nil {
			secretName := conn.Spec.Url.ValueFrom.SecretKeyRef.Name
			secretEntry, ok := secretsMap[secretName]
			if ok {
				secretId := secretEntry.Labels[k8s_utils.NeosyncUuidLabel]
				if connId != secretId {
					msg := fmt.Sprintf("connection and secret uuid mismatch. connId: %s secretId: %s", connId, secretId)
					return nil, nucleuserrors.NewInternalError(msg)
				}
				secret = secretEntry
			}
		}
		dto, err := dtomaps.ToConnectionDto(conn, secret)
		if err != nil {
			return nil, err
		}
		dtoConns = append(dtoConns, dto)
	}

	return connect.NewResponse(&mgmtv1alpha1.GetConnectionsByNamesResponse{
		Connections: dtoConns,
	}), nil
}

func (s *Service) GetConnection(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetConnectionRequest],
) (*connect.Response[mgmtv1alpha1.GetConnectionResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("connectionId", req.Msg.Id)

	connection, err := getConnectionById(ctx, logger, s.k8sclient, req.Msg.Id, s.cfg.JobConfigNamespace)
	if err != nil {
		return nil, err
	}

	dto, err := dtomaps.ToConnectionDto(connection.Connection, connection.Secret)
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
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("connectionName", req.Msg.Name)
	logger.Info("creating connection")
	connUuid := uuid.NewString()
	connectionString, err := getConnectionUrl(req.Msg.ConnectionConfig)
	if err != nil {
		return nil, err
	}

	connSecret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Msg.Name,
			Namespace: s.cfg.JobConfigNamespace,
			Labels: map[string]string{
				k8s_utils.NeosyncUuidLabel: connUuid,
			},
		},
		StringData: map[string]string{
			conn_utils.ConnectionSecretUrlField: connectionString,
		},
		Type: corev1.SecretTypeOpaque,
	}

	connection := &neosyncdevv1alpha1.SqlConnection{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: s.cfg.JobConfigNamespace,
			Name:      req.Msg.Name,
			Labels: map[string]string{
				k8s_utils.NeosyncUuidLabel: connUuid,
			},
		},
		Spec: neosyncdevv1alpha1.SqlConnectionSpec{
			Driver: neosyncdevv1alpha1.PostgresDriver,
			Url: neosyncdevv1alpha1.SqlConnectionUrl{
				ValueFrom: &neosyncdevv1alpha1.SqlConnectionUrlSource{
					SecretKeyRef: &neosyncdevv1alpha1.ConfigSelector{
						Name: connSecret.Name,
						Key:  "url",
					},
				},
			},
		},
	}

	secretChan := make(chan *corev1.Secret, 1)
	errs, errCtx := errgroup.WithContext(ctx)
	errs.Go(func() error {
		logger.Info("creating connection secret")
		createdSecret, err := s.k8sclient.K8sClient.CoreV1().Secrets(s.cfg.JobConfigNamespace).Create(errCtx, connSecret, metav1.CreateOptions{})
		if err != nil && !errors.IsAlreadyExists(err) {
			return err
		} else if err != nil && errors.IsAlreadyExists(err) {
			logger.Info("secret already exists, updating...")
			createdSecret, err = s.k8sclient.K8sClient.CoreV1().Secrets(s.cfg.JobConfigNamespace).Update(errCtx, connSecret, metav1.UpdateOptions{})
			if err != nil {
				logger.Error("unable to update connection secret")
				return err
			}
		}
		select {
		case secretChan <- createdSecret:
			return nil
		case <-errCtx.Done():
			return errCtx.Err()
		}
	})

	errs.Go(func() error {
		logger.Info("creating connection")
		err = s.k8sclient.CustomResourceClient.Create(errCtx, connection)
		if err != nil {
			return err
		}
		return nil
	})

	err = errs.Wait()
	if err != nil && !errors.IsAlreadyExists(err) {
		deleteSecretErr := s.k8sclient.K8sClient.CoreV1().Secrets(s.cfg.JobConfigNamespace).Delete(ctx, connSecret.Name, metav1.DeleteOptions{})
		if deleteSecretErr != nil {
			logger.Error("unable to clean up connection secret")
		}
		deleteConnErr := s.k8sclient.CustomResourceClient.Delete(ctx, connection, &runtimeclient.DeleteOptions{})
		if deleteConnErr != nil {
			logger.Error("unable to clean up connection")
		}
		return nil, err
	} else if err != nil && errors.IsAlreadyExists(err) {
		return nil, err
	}
	secret := <-secretChan
	close(secretChan)

	dto, err := dtomaps.ToConnectionDto(connection, secret)
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
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("connectionId", req.Msg.Id)
	logger.Info("updating connection")
	connection, err := getConnectionById(ctx, logger, s.k8sclient, req.Msg.Id, s.cfg.JobConfigNamespace)
	if err != nil {
		return nil, err
	}

	connectionString, err := getConnectionUrl(req.Msg.ConnectionConfig)
	if err != nil {
		return nil, err
	}

	var secret *corev1.Secret
	if connection.Secret != nil {
		logger.Info("updating secret")
		patch := &corev1.Secret{
			StringData: map[string]string{
				"url": connectionString,
			},
		}
		patchBits, err := json.Marshal(patch)
		if err != nil {
			return nil, err
		}
		secret, err = s.k8sclient.K8sClient.CoreV1().Secrets(s.cfg.JobConfigNamespace).Patch(
			ctx,
			connection.Secret.Name,
			types.MergePatchType,
			patchBits,
			metav1.PatchOptions{},
		)
		if err != nil {
			logger.Error("unable to update connection")
			return nil, err
		}
	} else if connection.Connection.Spec.Url.Value != nil && *connection.Connection.Spec.Url.Value != "" {
		logger.Info("updating connection url")
		patch := &neosyncdevv1alpha1.SqlConnection{
			Spec: neosyncdevv1alpha1.SqlConnectionSpec{
				Driver: neosyncdevv1alpha1.PostgresDriver,
				Url: neosyncdevv1alpha1.SqlConnectionUrl{
					Value: &connectionString,
				},
			},
		}
		patchBits, err := json.Marshal(patch)
		if err != nil {
			return nil, err
		}

		err = s.k8sclient.CustomResourceClient.Patch(ctx, connection.Connection, runtimeclient.RawPatch(types.MergePatchType, patchBits))
		if err != nil {
			logger.Error("unable to update connection")
			return nil, err
		}

	} else {
		return nil, nucleuserrors.NewNotImplemented("this connection config is not currently supported")
	}

	dto, err := dtomaps.ToConnectionDto(connection.Connection, secret)
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
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("connectionId", req.Msg.Id)
	logger.Info("deleting connection")

	conn, err := getSqlConnectionById(ctx, logger, s.k8sclient, req.Msg.Id, s.cfg.JobConfigNamespace)
	if err != nil && !nucleuserrors.IsNotFound(err) {
		return nil, err
	}

	secret, err := getConnectionSecretById(ctx, logger, s.k8sclient, req.Msg.Id, s.cfg.JobConfigNamespace)
	if err != nil && !nucleuserrors.IsNotFound(err) {
		return nil, err
	}

	errs, errCtx := errgroup.WithContext(ctx)
	if secret != nil {
		errs.Go(func() error {
			err := s.k8sclient.K8sClient.CoreV1().Secrets(s.cfg.JobConfigNamespace).Delete(errCtx, secret.Name, metav1.DeleteOptions{})
			if err != nil && !errors.IsNotFound(err) {
				return err
			} else if err != nil && errors.IsNotFound(err) {
				return nil
			}
			return nil
		})
	}

	if conn != nil {
		errs.Go(func() error {
			err := s.k8sclient.CustomResourceClient.Delete(errCtx, conn, &runtimeclient.DeleteOptions{})
			if err != nil && !errors.IsNotFound(err) {
				return err
			} else if err != nil && errors.IsNotFound(err) {
				return nil
			}
			return nil
		})
	}

	err = errs.Wait()
	if err != nil {
		logger.Error("unable to delete connection")
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.DeleteConnectionResponse{}), nil
}

func getConnectionUrl(c *mgmtv1alpha1.ConnectionConfig) (string, error) {
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

type connection struct {
	Connection *neosyncdevv1alpha1.SqlConnection
	Secret     *corev1.Secret
}

func getConnectionById(
	ctx context.Context,
	logger *slog.Logger,
	k8sclient *neosync_k8sclient.Client,
	id,
	namespace string,
) (*connection, error) {
	conn, err := getSqlConnectionById(ctx, logger, k8sclient, id, namespace)
	if err != nil {
		return nil, err
	}

	if conn.Spec.Url.ValueFrom != nil {
		secretName := conn.Spec.Url.ValueFrom.SecretKeyRef.Name
		secret, err := getConnectionSecretByName(ctx, logger, k8sclient, secretName, namespace)
		if err != nil {
			return nil, err
		}
		secretId := secret.Labels[k8s_utils.NeosyncUuidLabel]
		if conn.Labels[k8s_utils.NeosyncUuidLabel] != secretId {
			msg := fmt.Sprintf("connection and secret uuid mismatch. connId: %s secretId: %s", id, secretId)
			return nil, nucleuserrors.NewInternalError(msg)
		}
		return &connection{
			Connection: conn,
			Secret:     secret,
		}, nil
	}

	return &connection{
		Connection: conn,
	}, nil

}

func getConnectionSecretByName(
	ctx context.Context,
	logger *slog.Logger,
	k8sclient *neosync_k8sclient.Client,
	name string,
	namespace string,
) (*corev1.Secret, error) {
	secret, err := k8sclient.K8sClient.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	} else if err != nil && errors.IsNotFound(err) {
		logger.Error("connection secret not found", "secretName", name)
		return nil, err
	}
	return secret, nil
}

func getConnectionSecretById(
	ctx context.Context,
	logger *slog.Logger,
	k8sclient *neosync_k8sclient.Client,
	id string,
	namespace string,
) (*corev1.Secret, error) {
	req, err := labels.NewRequirement(k8s_utils.NeosyncUuidLabel, selection.Equals, []string{id})
	if err != nil {
		return nil, err
	}
	labelSelector := labels.NewSelector().Add(*req)
	secrets, err := k8sclient.K8sClient.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		logger.Error("unable to retrieve secrets")
		return nil, err
	}
	if len(secrets.Items) > 1 {
		return nil, nucleuserrors.NewInternalError(fmt.Sprintf("more than 1 secret found. id: %s", id))
	}
	if len(secrets.Items) == 0 {
		return nil, err
	}
	return &secrets.Items[0], nil
}

func getSqlConnectionById(
	ctx context.Context,
	logger *slog.Logger,
	k8sclient *neosync_k8sclient.Client,
	id string,
	namespace string,
) (*neosyncdevv1alpha1.SqlConnection, error) {
	conns := &neosyncdevv1alpha1.SqlConnectionList{}
	err := k8sclient.CustomResourceClient.List(ctx, conns, runtimeclient.InNamespace(namespace), &runtimeclient.MatchingLabels{
		k8s_utils.NeosyncUuidLabel: id,
	})
	if err != nil {
		logger.Error("unable to retrieve connection")
		return nil, err
	}
	if len(conns.Items) == 0 {
		return nil, nucleuserrors.NewNotFound(fmt.Sprintf("connection not found. id: %s", id))
	}
	if len(conns.Items) > 1 {
		return nil, nucleuserrors.NewInternalError(fmt.Sprintf("more than 1 connection found. id: %s", id))
	}
	return &conns.Items[0], nil
}
