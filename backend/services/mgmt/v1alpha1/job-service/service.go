package v1alpha1_jobservice

import (
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	clientmanager "github.com/nucleuscloud/neosync/backend/internal/temporal/client-manager"
)

type Service struct {
	cfg                *Config
	db                 *nucleusdb.NucleusDb
	connectionService  mgmtv1alpha1connect.ConnectionServiceClient
	useraccountService mgmtv1alpha1connect.UserAccountServiceClient

	temporalWfManager clientmanager.TemporalClientManagerClient
}

type Config struct {
	IsAuthEnabled           bool
	IsKubernetesEnabled     bool
	KubernetesNamespace     string
	KubernetesWorkerAppName string
}

func New(
	cfg *Config,
	db *nucleusdb.NucleusDb,
	temporalWfManager clientmanager.TemporalClientManagerClient,
	connectionService mgmtv1alpha1connect.ConnectionServiceClient,
	useraccountService mgmtv1alpha1connect.UserAccountServiceClient,
) *Service {
	return &Service{
		cfg:                cfg,
		db:                 db,
		temporalWfManager:  temporalWfManager,
		connectionService:  connectionService,
		useraccountService: useraccountService,
	}
}
