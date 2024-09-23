package v1alpha1_useraccountservice

import (
	auth_client "github.com/nucleuscloud/neosync/backend/internal/auth/client"
	"github.com/nucleuscloud/neosync/backend/internal/authmgmt"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
	clientmanager "github.com/nucleuscloud/neosync/backend/internal/temporal/client-manager"
	"github.com/nucleuscloud/neosync/internal/billing"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

type Service struct {
	cfg                   *Config
	db                    *neosyncdb.NeosyncDb
	temporalClientManager clientmanager.TemporalClientManagerClient
	authclient            auth_client.Interface
	authadminclient       authmgmt.Interface
	prometheusclient      promv1.API
	billingclient         billing.Interface
}

type Config struct {
	IsAuthEnabled            bool
	IsNeosyncCloud           bool
	DefaultMaxAllowedRecords *int64
}

func New(
	cfg *Config,
	db *neosyncdb.NeosyncDb,
	temporalClientManager clientmanager.TemporalClientManagerClient,
	authclient auth_client.Interface,
	authadminclient authmgmt.Interface,
	prometheusclient promv1.API,
	billingclient billing.Interface,
) *Service {
	return &Service{
		cfg:                   cfg,
		db:                    db,
		temporalClientManager: temporalClientManager,
		authclient:            authclient,
		authadminclient:       authadminclient,
		prometheusclient:      prometheusclient,
		billingclient:         billingclient,
	}
}
