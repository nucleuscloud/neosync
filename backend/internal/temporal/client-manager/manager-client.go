package clientmanager

import (
	"context"
	"crypto/tls"
	"errors"
	"log/slog"
	"sync"

	temporalclient "go.temporal.io/sdk/client"

	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	jsonmodels "github.com/nucleuscloud/neosync/backend/internal/nucleusdb/json-models"
)

type TemporalClientManager struct {
	config      *Config
	db          DB
	dbtx        db_queries.DBTX
	wfmap       *sync.Map
	accountWfMu *sync.Map

	nsmap       *sync.Map
	accountNsMu *sync.Map
}

type TemporalClientManagerClient interface {
	ClearNamespaceClientByAccount(ctx context.Context, accountId string)
	ClearWorkflowClientByAccount(ctx context.Context, accountId string)
	GetNamespaceClientByAccount(ctx context.Context, accountId string, logger *slog.Logger) (temporalclient.NamespaceClient, error)
	GetWorkflowClientByAccount(ctx context.Context, accountId string, logger *slog.Logger) (temporalclient.Client, error)
	GetScheduleClientByAccount(ctx context.Context, accountId string, logger *slog.Logger) (temporalclient.ScheduleClient, error)
	GetScheduleHandleClientByAccount(ctx context.Context, accountId string, scheduleId string, logger *slog.Logger) (temporalclient.ScheduleHandle, error)
}

type DB interface {
	GetTemporalConfigByAccount(ctx context.Context, db db_queries.DBTX, accountId pgtype.UUID) (*jsonmodels.TemporalConfig, error)
}

type Config struct {
	AuthCertificates []tls.Certificate
}

func New(
	config *Config,
	db DB,
	dbtx db_queries.DBTX,
) *TemporalClientManager {
	return &TemporalClientManager{
		config:      config,
		db:          db,
		wfmap:       &sync.Map{},
		accountWfMu: &sync.Map{},
		nsmap:       &sync.Map{},
		accountNsMu: &sync.Map{},
		dbtx:        dbtx,
	}
}

func (t *TemporalClientManager) ClearNamespaceClientByAccount(
	ctx context.Context,
	accountId string,
) {
	client, ok := t.loadNsClientByAccount(accountId)
	if ok {
		defer client.Close()
		t.nsmap.Delete(accountId)
	}
}

func (t *TemporalClientManager) ClearWorkflowClientByAccount(
	ctx context.Context,
	accountId string,
) {
	client, ok := t.loadWfClientByAccount(accountId)
	if ok {
		defer client.Close()
		t.wfmap.Delete(accountId)
	}
}

func (t *TemporalClientManager) loadWfClientByAccount(accountId string) (temporalclient.Client, bool) {
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

func (t *TemporalClientManager) loadNsClientByAccount(accountId string) (temporalclient.NamespaceClient, bool) {
	client, ok := t.nsmap.Load(accountId)
	if ok {
		tclient, ok := client.(temporalclient.NamespaceClient)
		if !ok {
			return nil, false
		}
		return tclient, true
	}
	return nil, false
}

func (t *TemporalClientManager) GetNamespaceClientByAccount(
	ctx context.Context,
	accountId string,
	logger *slog.Logger,
) (temporalclient.NamespaceClient, error) {
	client, ok := t.loadNsClientByAccount(accountId)
	if ok {
		return client, nil
	}

	mu, _ := t.accountNsMu.LoadOrStore(accountId, &sync.Mutex{})
	mutex := mu.(*sync.Mutex)
	mutex.Lock()
	defer mutex.Unlock()

	client, ok = t.loadNsClientByAccount(accountId)
	if ok {
		return client, nil
	}
	client, err := t.getNewNSClientByAccount(ctx, accountId, logger)
	if err != nil {
		return nil, err
	}
	t.nsmap.Store(accountId, client)
	return client, nil
}

func (t *TemporalClientManager) GetScheduleClientByAccount(
	ctx context.Context,
	accountId string,
	logger *slog.Logger,
) (temporalclient.ScheduleClient, error) {
	client, err := t.GetWorkflowClientByAccount(ctx, accountId, logger)
	if err != nil {
		return nil, err
	}
	return client.ScheduleClient(), nil
}

func (t *TemporalClientManager) GetScheduleHandleClientByAccount(
	ctx context.Context,
	accountId string,
	scheduleId string,
	logger *slog.Logger,
) (temporalclient.ScheduleHandle, error) {
	client, err := t.GetScheduleClientByAccount(ctx, accountId, logger)
	if err != nil {
		return nil, err
	}
	return client.GetHandle(ctx, scheduleId), nil
}

func (t *TemporalClientManager) GetWorkflowClientByAccount(
	ctx context.Context,
	accountId string,
	logger *slog.Logger,
) (temporalclient.Client, error) {
	client, ok := t.loadWfClientByAccount(accountId)
	if ok {
		return client, nil
	}

	mu, _ := t.accountWfMu.LoadOrStore(accountId, &sync.Mutex{})
	mutex := mu.(*sync.Mutex)
	mutex.Lock()
	defer mutex.Unlock()

	client, ok = t.loadWfClientByAccount(accountId)
	if ok {
		return client, nil
	}
	client, err := t.getNewWFClientByAccount(ctx, accountId, logger)
	if err != nil {
		return nil, err
	}
	t.wfmap.Store(accountId, client)
	return client, nil
}

func (t *TemporalClientManager) getNewNSClientByAccount(
	ctx context.Context,
	accountId string,
	logger *slog.Logger,
) (temporalclient.NamespaceClient, error) {
	tc, err := t.getTemporalConfigByAccount(ctx, accountId)
	if err != nil {
		return nil, err
	}
	if tc.Namespace == "" {
		return nil, errors.New("neosync account does not have a configured temporal namespace")
	}

	return temporalclient.NewNamespaceClient(*t.getClientOptions(accountId, tc, logger))
}

func (t *TemporalClientManager) getNewWFClientByAccount(
	ctx context.Context,
	accountId string,
	logger *slog.Logger,
) (temporalclient.Client, error) {
	tc, err := t.getTemporalConfigByAccount(ctx, accountId)
	if err != nil {
		return nil, err
	}
	if tc.Namespace == "" {
		return nil, errors.New("neosync account does not have a configured temporal namespace")
	}
	return temporalclient.NewLazyClient(*t.getClientOptions(accountId, tc, logger))
}

func (t *TemporalClientManager) getTemporalConfigByAccount(
	ctx context.Context,
	accountId string,
) (*jsonmodels.TemporalConfig, error) {
	accountUuid, err := nucleusdb.ToUuid(accountId)
	if err != nil {
		return nil, err
	}
	return t.db.GetTemporalConfigByAccount(ctx, t.dbtx, accountUuid)
}

func (t *TemporalClientManager) getClientOptions(
	accountId string,
	tc *jsonmodels.TemporalConfig,
	logger *slog.Logger,
) *temporalclient.Options {
	opts := temporalclient.Options{
		Logger: logger.With(
			"temporal-client", "true",
			"neosync-account-id", accountId,
		),
		HostPort:  tc.Url,
		Namespace: tc.Namespace,
	}
	connectOpts := t.getClientConnectionOptions()
	if connectOpts != nil {
		opts.ConnectionOptions = *connectOpts
	}
	return &opts
}

func (t *TemporalClientManager) getClientConnectionOptions() *temporalclient.ConnectionOptions {
	if len(t.config.AuthCertificates) == 0 {
		return nil
	}
	return &temporalclient.ConnectionOptions{
		TLS: &tls.Config{
			Certificates: t.config.AuthCertificates,
			MinVersion:   tls.VersionTLS13,
		},
	}
}
