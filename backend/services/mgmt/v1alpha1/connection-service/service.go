package v1alpha1_connectionservice

import (
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
	"github.com/nucleuscloud/neosync/backend/internal/userdata"
	"github.com/nucleuscloud/neosync/backend/pkg/mongoconnect"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	awsmanager "github.com/nucleuscloud/neosync/internal/aws"
)

type Service struct {
	cfg            *Config
	db             *neosyncdb.NeosyncDb
	userclient     userdata.Interface
	sqlConnector   sqlconnect.SqlConnector
	sqlmanager     sql_manager.SqlManagerClient
	mongoconnector mongoconnect.Interface
	awsManager     awsmanager.NeosyncAwsManagerClient
}

type Config struct {
}

func New(
	cfg *Config,
	db *neosyncdb.NeosyncDb,
	userclient userdata.Interface,
	mongoconnector mongoconnect.Interface,
	awsManager awsmanager.NeosyncAwsManagerClient,
	sqlmanager sql_manager.SqlManagerClient,
	sqlconnector sqlconnect.SqlConnector,
) *Service {
	return &Service{
		cfg:            cfg,
		db:             db,
		userclient:     userclient,
		sqlmanager:     sqlmanager,
		mongoconnector: mongoconnector,
		awsManager:     awsManager,
		sqlConnector:   sqlconnector,
	}
}
