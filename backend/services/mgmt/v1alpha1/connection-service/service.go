package v1alpha1_connectionservice

import (
	neosync_k8sclient "github.com/nucleuscloud/neosync/backend/internal/k8s/client"
)

type Service struct {
	cfg       *Config
	k8sclient *neosync_k8sclient.Client
}

type Config struct {
	JobConfigNamespace string
}

func New(
	cfg *Config,
	k8sclient *neosync_k8sclient.Client,
) *Service {

	return &Service{
		cfg:       cfg,
		k8sclient: k8sclient,
	}
}
