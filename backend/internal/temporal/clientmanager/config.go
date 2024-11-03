package clientmanager

import (
	"crypto/tls"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
)

type TemporalConfig struct {
	Url              string
	Namespace        string
	SyncJobQueueName string
	TLSConfig        *tls.Config
}

// Combines default and account-specific configs, with account taking precedence
func (c *TemporalConfig) Override(config *TemporalConfig) *TemporalConfig {
	result := *c
	if config == nil {
		config = &TemporalConfig{}
	}
	if config.Url != "" {
		result.Url = config.Url
	}
	if config.Namespace != "" {
		result.Namespace = config.Namespace
	}
	if config.SyncJobQueueName != "" {
		result.SyncJobQueueName = config.SyncJobQueueName
	}
	if config.TLSConfig != nil {
		result.TLSConfig = config.TLSConfig
	}
	return &result
}

func (c *TemporalConfig) Equals(other *TemporalConfig) bool {
	return c.Url == other.Url &&
		c.Namespace == other.Namespace &&
		c.SyncJobQueueName == other.SyncJobQueueName
}

func (c *TemporalConfig) ToDto() *mgmtv1alpha1.AccountTemporalConfig {
	return &mgmtv1alpha1.AccountTemporalConfig{
		Url:              c.Url,
		Namespace:        c.Namespace,
		SyncJobQueueName: c.SyncJobQueueName,
	}
}
