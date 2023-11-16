package v1alpha1_useraccountservice

import (
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
)

type Service struct {
	cfg         *Config
	db          *nucleusdb.NucleusDb
	authService mgmtv1alpha1connect.AuthServiceClient
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
	authService mgmtv1alpha1connect.AuthServiceClient,
) *Service {

	return &Service{
		cfg:         cfg,
		db:          db,
		authService: authService,
	}
}
