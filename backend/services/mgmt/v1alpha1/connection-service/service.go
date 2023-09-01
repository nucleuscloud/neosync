package v1alpha1_connectionservice

import (
	neosync_k8sclient "github.com/nucleuscloud/neosync/backend/internal/k8s/client"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
)

type Service struct {
	cfg       *Config
	db        *nucleusdb.NucleusDb
	k8sclient *neosync_k8sclient.Client
}

type Config struct {
	JobConfigNamespace string
}

func New(
	cfg *Config,
	db *nucleusdb.NucleusDb,
	k8sclient *neosync_k8sclient.Client,
) *Service {

	return &Service{
		cfg:       cfg,
		db:        db,
		k8sclient: k8sclient,
	}
}
