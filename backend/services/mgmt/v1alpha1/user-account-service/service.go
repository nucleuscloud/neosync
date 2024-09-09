package v1alpha1_useraccountservice

import (
	auth_client "github.com/nucleuscloud/neosync/backend/internal/auth/client"
	"github.com/nucleuscloud/neosync/backend/internal/authmgmt"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	clientmanager "github.com/nucleuscloud/neosync/backend/internal/temporal/client-manager"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

type Service struct {
	cfg                   *Config
	db                    *nucleusdb.NucleusDb
	temporalClientManager clientmanager.TemporalClientManagerClient
	authclient            auth_client.Interface
	authadminclient       authmgmt.Interface
	prometheusclient      promv1.API
}

type Config struct {
	IsAuthEnabled  bool
	IsNeosyncCloud bool
}

func New(
	cfg *Config,
	db *nucleusdb.NucleusDb,
	temporalClientManager clientmanager.TemporalClientManagerClient,
	authclient auth_client.Interface,
	authadminclient authmgmt.Interface,
	prometheusclient promv1.API,
) *Service {
	return &Service{
		cfg:                   cfg,
		db:                    db,
		temporalClientManager: temporalClientManager,
		authclient:            authclient,
		authadminclient:       authadminclient,
		prometheusclient:      prometheusclient,
	}
}
