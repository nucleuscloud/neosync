package v1alpha1_useraccountservice

import (
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
)

type Service struct {
	cfg *Config
	db  *nucleusdb.NucleusDb
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
) *Service {

	return &Service{
		cfg: cfg,
		db:  db,
	}
}
