package v1alpha1_jobservice

import (
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	jobhooks "github.com/nucleuscloud/neosync/backend/internal/ee/hooks/jobs"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
	clientmanager "github.com/nucleuscloud/neosync/backend/internal/temporal/clientmanager"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	"github.com/nucleuscloud/neosync/internal/ee/rbac"
)

type Service struct {
	cfg                *Config
	db                 *neosyncdb.NeosyncDb
	connectionService  mgmtv1alpha1connect.ConnectionServiceClient
	useraccountService mgmtv1alpha1connect.UserAccountServiceClient
	sqlmanager         sql_manager.SqlManagerClient

	temporalmgr clientmanager.Interface

	hookService jobhooks.Interface

	rbacClient *rbac.Rbac
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

	LabelsQuery string // Labels to filter loki by, without the curly braces
	// labels to keep after the json filtering. Keeps ordering. Not sure if it will always need to equal the labels query keys, so separating this
	KeepLabels []string
}

type Config struct {
	IsAuthEnabled  bool
	IsNeosyncCloud bool

	RunLogConfig *RunLogConfig
}

type RunLogConfig struct {
	IsEnabled        bool
	RunLogType       *RunLogType
	RunLogPodConfig  *KubePodRunLogConfig // required if RunLogType is k8s-pods
	LokiRunLogConfig *LokiRunLogConfig    // required if RunLogType is loki
}

func New(
	cfg *Config,
	db *neosyncdb.NeosyncDb,
	temporalWfManager clientmanager.Interface,
	connectionService mgmtv1alpha1connect.ConnectionServiceClient,
	useraccountService mgmtv1alpha1connect.UserAccountServiceClient,
	sqlmanager sql_manager.SqlManagerClient,
	jobhookService jobhooks.Interface,
	rbacClient *rbac.Rbac,
) *Service {
	return &Service{
		cfg:                cfg,
		db:                 db,
		temporalmgr:        temporalWfManager,
		connectionService:  connectionService,
		useraccountService: useraccountService,
		sqlmanager:         sqlmanager,
		hookService:        jobhookService,
		rbacClient:         rbacClient,
	}
}
