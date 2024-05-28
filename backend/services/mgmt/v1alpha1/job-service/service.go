package v1alpha1_jobservice

import (
	"sync"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	clientmanager "github.com/nucleuscloud/neosync/backend/internal/temporal/client-manager"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
)

type Service struct {
	cfg                *Config
	db                 *nucleusdb.NucleusDb
	connectionService  mgmtv1alpha1connect.ConnectionServiceClient
	useraccountService mgmtv1alpha1connect.UserAccountServiceClient
	sqlmanager         sql_manager.SqlManagerClient

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

	LabelsQuery string // Labels to filter loki by, without the curly braces
	// labels to keep after the json filtering. Keeps ordering. Not sure if it will always need to equal the labels query keys, so separating this
	KeepLabels []string
}

type Config struct {
	IsAuthEnabled bool

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
	db *nucleusdb.NucleusDb,
	temporalWfManager clientmanager.TemporalClientManagerClient,
	connectionService mgmtv1alpha1connect.ConnectionServiceClient,
	useraccountService mgmtv1alpha1connect.UserAccountServiceClient,
	sqlConnector sqlconnect.SqlConnector,
	pgquerier pg_queries.Querier,
	mysqlquerier mysql_queries.Querier,
) *Service {
	pgpoolmap := &sync.Map{}
	mysqlpoolmap := &sync.Map{}

	sqlmanager := sql_manager.NewSqlManager(pgpoolmap, pgquerier, mysqlpoolmap, mysqlquerier, sqlConnector)
	return &Service{
		cfg:                cfg,
		db:                 db,
		temporalWfManager:  temporalWfManager,
		connectionService:  connectionService,
		useraccountService: useraccountService,
		sqlmanager:         sqlmanager,
	}
}
