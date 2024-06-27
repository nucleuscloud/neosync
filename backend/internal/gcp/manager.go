package neosync_gcp

import (
	"context"
	"log/slog"

	"cloud.google.com/go/storage"
)

type ManagerInterface interface {
	GetClient(ctx context.Context, logger *slog.Logger) (ClientInterface, error)
}

type Manager struct{}

var _ ManagerInterface = &Manager{}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) GetClient(ctx context.Context, logger *slog.Logger) (ClientInterface, error) {
	sc, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return NewClient(sc, logger), nil
}
