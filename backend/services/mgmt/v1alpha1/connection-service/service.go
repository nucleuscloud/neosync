package v1alpha1_connectionservice

import (
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
	"github.com/nucleuscloud/neosync/backend/pkg/mongoconnect"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	awsmanager "github.com/nucleuscloud/neosync/internal/aws"
)

type Service struct {
	cfg                *Config
	db                 *neosyncdb.NeosyncDb
	useraccountService mgmtv1alpha1connect.UserAccountServiceClient
	sqlConnector       sqlconnect.SqlConnector
	sqlmanager         sql_manager.SqlManagerClient
	mongoconnector     mongoconnect.Interface
	awsManager         awsmanager.NeosyncAwsManagerClient
}

type Config struct {
}

func New(
	cfg *Config,
	db *neosyncdb.NeosyncDb,
	useraccountService mgmtv1alpha1connect.UserAccountServiceClient,
	mongoconnector mongoconnect.Interface,
	awsManager awsmanager.NeosyncAwsManagerClient,
	sqlmanager sql_manager.SqlManagerClient,
	sqlconnector sqlconnect.SqlConnector,
) *Service {
	return &Service{
		cfg:                cfg,
		db:                 db,
		useraccountService: useraccountService,
		sqlmanager:         sqlmanager,
		mongoconnector:     mongoconnector,
		awsManager:         awsManager,
		sqlConnector:       sqlconnector,
	}
}
