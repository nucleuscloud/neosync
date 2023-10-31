package v1alpha1_jobservice

import (
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	workflowmanager "github.com/nucleuscloud/neosync/backend/internal/temporal/workflow-manager"
	v1alpha1_connectionservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/connection-service"
	v1alpha1_useraccountservice "github.com/nucleuscloud/neosync/backend/services/mgmt/v1alpha1/user-account-service"
	temporalclient "go.temporal.io/sdk/client"
)

type NameGenerator interface {
	Generate() string
}

type Service struct {
	cfg                *Config
	db                 *nucleusdb.NucleusDb
	temporalNsClient   temporalclient.NamespaceClient
	connectionService  *v1alpha1_connectionservice.Service
	useraccountService *v1alpha1_useraccountservice.Service

	namegenerator NameGenerator

	temporalWfManager *workflowmanager.TemporalWorkflowManager
}

type Config struct {
	IsAuthEnabled bool
}

func New(
	cfg *Config,
	db *nucleusdb.NucleusDb,
	temporalNsClient temporalclient.NamespaceClient,
	temporalWfManager *workflowmanager.TemporalWorkflowManager,
	namegenerator NameGenerator,
	connectionService *v1alpha1_connectionservice.Service,
	useraccountService *v1alpha1_useraccountservice.Service,
) *Service {

	return &Service{
		cfg:                cfg,
		db:                 db,
		temporalNsClient:   temporalNsClient,
		temporalWfManager:  temporalWfManager,
		namegenerator:      namegenerator,
		connectionService:  connectionService,
		useraccountService: useraccountService,
	}
}
