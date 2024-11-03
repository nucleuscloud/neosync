package clientmanager

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
	pg_models "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"
)

type ConfigProvider interface {
	GetConfig(ctx context.Context, accountID string) (*TemporalConfig, error)
}

type DB interface {
	GetTemporalConfigByAccount(ctx context.Context, db db_queries.DBTX, accountId pgtype.UUID) (*pg_models.TemporalConfig, error)
}

type DBConfigProvider struct {
	defaultConfig *TemporalConfig
	db            DB
	dbtx          db_queries.DBTX
}

func NewDBConfigProvider(defaultConfig *TemporalConfig, db DB, dbtx db_queries.DBTX) *DBConfigProvider {
	defaultConfig.isDefault = true
	return &DBConfigProvider{
		defaultConfig: defaultConfig,
		db:            db,
		dbtx:          dbtx,
	}
}

func (p *DBConfigProvider) GetConfig(ctx context.Context, accountID string) (*TemporalConfig, error) {
	accountUuid, err := neosyncdb.ToUuid(accountID)
	if err != nil {
		return nil, fmt.Errorf("invalid account ID: %w", err)
	}

	dbConfig, err := p.db.GetTemporalConfigByAccount(ctx, p.dbtx, accountUuid)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve temporal config: %w", err)
	}

	// If the account has no specific configuration, return the default
	if dbConfig.Url == "" && dbConfig.Namespace == "" && dbConfig.SyncJobQueueName == "" {
		return p.defaultConfig, nil
	}

	// Otherwise, merge with defaults and mark as non-default
	accountConfig := dbConfigToTemporalConfig(dbConfig)
	mergedConfig := p.defaultConfig.Override(accountConfig)
	mergedConfig.isDefault = false
	return mergedConfig, nil
}

func dbConfigToTemporalConfig(dbConfig *pg_models.TemporalConfig) *TemporalConfig {
	return &TemporalConfig{
		Url:              dbConfig.Url,
		Namespace:        dbConfig.Namespace,
		SyncJobQueueName: dbConfig.SyncJobQueueName,
	}
}
