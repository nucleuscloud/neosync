package clientmanager

import (
	"context"
	"log/slog"

	temporalclient "go.temporal.io/sdk/client"
)

type ClientFactory interface {
	CreateNamespaceClient(ctx context.Context, config *TemporalConfig, logger *slog.Logger) (temporalclient.NamespaceClient, error)
	CreateWorkflowClient(ctx context.Context, config *TemporalConfig, logger *slog.Logger) (temporalclient.Client, error)
}

type TemporalClientFactory struct{}

var _ ClientFactory = (*TemporalClientFactory)(nil)

func NewTemporalClientFactory() *TemporalClientFactory {
	return &TemporalClientFactory{}
}

func (f *TemporalClientFactory) CreateNamespaceClient(
	ctx context.Context,
	config *TemporalConfig,
	logger *slog.Logger,
) (temporalclient.NamespaceClient, error) {
	return temporalclient.NewNamespaceClient(f.getClientOptions(config, logger))
}

func (f *TemporalClientFactory) CreateWorkflowClient(
	ctx context.Context,
	config *TemporalConfig,
	logger *slog.Logger,
) (temporalclient.Client, error) {
	return temporalclient.NewLazyClient(f.getClientOptions(config, logger))
}

func (f *TemporalClientFactory) getClientOptions(config *TemporalConfig, logger *slog.Logger) temporalclient.Options {
	opts := temporalclient.Options{
		Logger:    logger,
		HostPort:  config.Url,
		Namespace: config.Namespace,
	}

	if config.TLSConfig != nil {
		opts.ConnectionOptions = temporalclient.ConnectionOptions{
			TLS: config.TLSConfig,
		}
	}

	return opts
}
