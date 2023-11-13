package v1alpha1_useraccountservice

import (
	authmgmt "github.com/nucleuscloud/neosync/backend/internal/auth-mgmt"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
)

type Service struct {
	cfg      *Config
	db       *nucleusdb.NucleusDb
	authMgmt *authmgmt.Auth0MgmtClient
}

type Config struct {
	IsAuthEnabled bool
	Temporal      *TemporalConfig
}

type TemporalConfig struct {
	DefaultTemporalNamespace        string
	DefaultTemporalSyncJobQueueName string
	DefaultTemporalUrl              string
}

func New(
	cfg *Config,
	db *nucleusdb.NucleusDb,
	authMgmt *authmgmt.Auth0MgmtClient,
) *Service {

	return &Service{
		cfg:      cfg,
		db:       db,
		authMgmt: authMgmt,
	}
}
