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

type RunLogType string

const (
	KubePodRunLogType RunLogType = "k8s-pods"
	LokiRunLogType    RunLogType = "loki"
)

type KubePodRunLogConfig struct {
	Namespace     string
	WorkerAppName string
}

type LokiRunLogConfig struct {
	BaseUrl string
}

type Config struct {
	IsAuthEnabled bool
	RunLogType    *RunLogType

	RunLogPodConfig  *KubePodRunLogConfig // required if RunLogType is k8s-pods
	LokiRunLogConfig *LokiRunLogConfig    // required if RunLogType is loki
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
