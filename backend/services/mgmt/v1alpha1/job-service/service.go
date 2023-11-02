package v1alpha1_jobservice

import (
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	clientmanager "github.com/nucleuscloud/neosync/backend/internal/temporal/client-manager"
	v1alpha1_connectionservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/connection-service"
	v1alpha1_useraccountservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/user-account-service"
)

type Service struct {
	cfg                *Config
	db                 *nucleusdb.NucleusDb
	connectionService  *v1alpha1_connectionservice.Service
	useraccountService *v1alpha1_useraccountservice.Service

	temporalWfManager *clientmanager.TemporalClientManager
}

type Config struct {
	IsAuthEnabled bool
}

func New(
	cfg *Config,
	db *nucleusdb.NucleusDb,
	temporalWfManager *clientmanager.TemporalClientManager,
	connectionService *v1alpha1_connectionservice.Service,
	useraccountService *v1alpha1_useraccountservice.Service,
) *Service {

	return &Service{
		cfg:                cfg,
		db:                 db,
		temporalWfManager:  temporalWfManager,
		connectionService:  connectionService,
		useraccountService: useraccountService,
	}
}
