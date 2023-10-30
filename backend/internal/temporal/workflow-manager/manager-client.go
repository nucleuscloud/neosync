package workflowmanager

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	temporalclient "go.temporal.io/sdk/client"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	jsonmodels "github.com/nucleuscloud/neosync/backend/internal/nucleusdb/json-models"
)

type TemporalWorkflowManager struct {
	config *Config
	db     DB
	wfmap  *sync.Map
}

type DB interface {
	GetTemporalConfigByAccount(ctx context.Context, accountId pgtype.UUID) (*jsonmodels.TemporalConfig, error)
}

type Config struct {
	TemporalUrl string
}

func New(
	config *Config,
	db DB,
) *TemporalWorkflowManager {
	return &TemporalWorkflowManager{
		config: config,
		db:     db,
		wfmap:  &sync.Map{},
	}
}

func (t *TemporalWorkflowManager) ClearClientByAccount(
	ctx context.Context,
	accountId string,
) {
	client, ok := t.loadClientByAccount(accountId)
	if ok {
		defer client.Close()
		t.wfmap.Delete(accountId)
	}
}

func (t *TemporalWorkflowManager) loadClientByAccount(accountId string) (temporalclient.Client, bool) {
	client, ok := t.wfmap.Load(accountId)
	if ok {
		tclient, ok := client.(temporalclient.Client)
		if !ok {
			return nil, false
		}
		return tclient, true
	}
	return nil, false
}

func (t *TemporalWorkflowManager) GetClientByAccount(
	ctx context.Context,
	accountId string,
	logger *slog.Logger,
) (temporalclient.Client, error) {
	client, ok := t.loadClientByAccount(accountId)
	if ok {
		return client, nil
	}

	client, err := t.getNewClientByAccount(ctx, accountId, logger)
	if err != nil {
		return nil, err
	}
	t.wfmap.Store(accountId, client)
	return client, nil
}

func (t *TemporalWorkflowManager) getNewClientByAccount(
	ctx context.Context,
	accountId string,
	logger *slog.Logger,
) (temporalclient.Client, error) {
	accountUuid, err := nucleusdb.ToUuid(accountId)
	if err != nil {
		return nil, err
	}
	tc, err := t.db.GetTemporalConfigByAccount(ctx, accountUuid)
	if err != nil {
		return nil, err
	}
	if tc.Namespace == "" {
		return nil, errors.New("neosync account does not have a configured temporal namespace")
	}

	return temporalclient.NewLazyClient(temporalclient.Options{
		Logger: logger.With(
			"temporal-client", "true",
			"neosync-account-id", accountId,
		),
		HostPort:  t.config.TemporalUrl,
		Namespace: tc.Namespace,
	})
}
