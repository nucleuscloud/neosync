package v1alpha1_useraccountservice

import (
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	clientmanager "github.com/nucleuscloud/neosync/backend/internal/temporal/client-manager"
)

type Service struct {
	cfg                   *Config
	db                    *nucleusdb.NucleusDb
	authService           mgmtv1alpha1connect.AuthServiceClient
	temporalClientManager clientmanager.TemporalClientManagerClient
}

type Config struct {
	IsAuthEnabled bool
}

func New(
	cfg *Config,
	db *nucleusdb.NucleusDb,
	authService mgmtv1alpha1connect.AuthServiceClient,
	temporalClientManager clientmanager.TemporalClientManagerClient,
) *Service {
	return &Service{
		cfg:                   cfg,
		db:                    db,
		authService:           authService,
		temporalClientManager: temporalClientManager,
	}
}
