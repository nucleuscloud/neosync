package v1alpha1_useraccountservice

import (
	"github.com/nucleuscloud/neosync/backend/internal/authmgmt/auth0"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	clientmanager "github.com/nucleuscloud/neosync/backend/internal/temporal/client-manager"
)

type Service struct {
	cfg                   *Config
	db                    *nucleusdb.NucleusDb
	auth0MgmtClient       auth0.Auth0MgmtClientInterface
	temporalClientManager clientmanager.TemporalClientManagerClient
}

type Config struct {
	IsAuthEnabled bool
}

func New(
	cfg *Config,
	db *nucleusdb.NucleusDb,
	auth0MgmtClient auth0.Auth0MgmtClientInterface,
	temporalClientManager clientmanager.TemporalClientManagerClient,
) *Service {
	return &Service{
		cfg:                   cfg,
		db:                    db,
		auth0MgmtClient:       auth0MgmtClient,
		temporalClientManager: temporalClientManager,
	}
}
