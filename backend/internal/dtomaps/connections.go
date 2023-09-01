package dtomaps

import (
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	conn_utils "github.com/nucleuscloud/neosync/backend/internal/utils/connections"
	k8s_utils "github.com/nucleuscloud/neosync/backend/internal/utils/k8s"
	neosyncdevv1alpha1 "github.com/nucleuscloud/neosync/k8s-operator/api/v1alpha1"
	"google.golang.org/protobuf/types/known/timestamppb"
	corev1 "k8s.io/api/core/v1"
)

func ToConnectionDto(
	input *neosyncdevv1alpha1.SqlConnection,
	secret *corev1.Secret,
) (*mgmtv1alpha1.Connection, error) {

	connectionConfig, err := getConnectionConfig(input, secret)
	if err != nil {
		return nil, err
	}

	labels := input.GetLabels()
	return &mgmtv1alpha1.Connection{
		Id:               labels[k8s_utils.NeosyncUuidLabel],
		Name:             input.Name,
		ConnectionConfig: connectionConfig,
		CreatedAt:        timestamppb.New(input.CreationTimestamp.Time),
		UpdatedAt:        timestamppb.New(input.CreationTimestamp.Time), // TODO
	}, nil
}

func getConnectionConfig(
	input *neosyncdevv1alpha1.SqlConnection,
	secret *corev1.Secret,
) (*mgmtv1alpha1.ConnectionConfig, error) {

	switch input.Spec.Driver {
	case neosyncdevv1alpha1.PostgresDriver:
		var url string
		if secret != nil {
			url = string(secret.Data["url"])
		} else if input.Spec.Url.Value != nil && *input.Spec.Url.Value != "" {
			url = *input.Spec.Url.Value
		}
		connCfg, err := conn_utils.ParsePostgresUrl(url)
		if err != nil {
			return nil, err
		}
		return &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
				PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
					ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Connection{
						Connection: &mgmtv1alpha1.PostgresConnection{
							Host:    connCfg.Host,
							Port:    connCfg.Port,
							Name:    connCfg.Database,
							User:    connCfg.User,
							Pass:    connCfg.Pass,
							SslMode: connCfg.SslMode,
						},
					},
				},
			},
		}, nil

	default:
		return nil, nucleuserrors.NewNotImplemented("this connection config is not currently supported")
	}
}
