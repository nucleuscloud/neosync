package v1alpha1_jobservice

import (
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	jobhooks "github.com/nucleuscloud/neosync/backend/internal/ee/hooks/jobs"
	"github.com/nucleuscloud/neosync/backend/internal/ee/rbac"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
	clientmanager "github.com/nucleuscloud/neosync/backend/internal/temporal/clientmanager"
	"github.com/nucleuscloud/neosync/backend/internal/userdata"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
)

type Service struct {
	cfg               *Config
	db                *neosyncdb.NeosyncDb
	connectionService mgmtv1alpha1connect.ConnectionServiceClient
	userdataclient    userdata.Interface
	sqlmanager        sql_manager.SqlManagerClient

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
	sqlmanager sql_manager.SqlManagerClient,
	jobhookService jobhooks.Interface,
	userdataclient userdata.Interface,
) *Service {
	return &Service{
		cfg:               cfg,
		db:                db,
		temporalmgr:       temporalWfManager,
		connectionService: connectionService,
		sqlmanager:        sqlmanager,
		hookService:       jobhookService,
		userdataclient:    userdataclient,
	}
}
