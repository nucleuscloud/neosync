package v1alpha1_useraccountservice

import (
	auth_client "github.com/nucleuscloud/neosync/backend/internal/auth/client"
	"github.com/nucleuscloud/neosync/backend/internal/authmgmt"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	clientmanager "github.com/nucleuscloud/neosync/backend/internal/temporal/client-manager"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	stripeapiclient "github.com/stripe/stripe-go/v79/client"
)

type Service struct {
	cfg                   *Config
	db                    *nucleusdb.NucleusDb
	temporalClientManager clientmanager.TemporalClientManagerClient
	authclient            auth_client.Interface
	authadminclient       authmgmt.Interface
	prometheusclient      promv1.API
	stripeclient          *stripeapiclient.API
}

type Config struct {
	IsAuthEnabled            bool
	IsNeosyncCloud           bool
	DefaultMaxAllowedRecords *int64
}

func New(
	cfg *Config,
	db *nucleusdb.NucleusDb,
	temporalClientManager clientmanager.TemporalClientManagerClient,
	authclient auth_client.Interface,
	authadminclient authmgmt.Interface,
	prometheusclient promv1.API,
	stripeclient *stripeapiclient.API,
) *Service {
	return &Service{
		cfg:                   cfg,
		db:                    db,
		temporalClientManager: temporalClientManager,
		authclient:            authclient,
		authadminclient:       authadminclient,
		prometheusclient:      prometheusclient,
		stripeclient:          stripeclient,
	}
}
