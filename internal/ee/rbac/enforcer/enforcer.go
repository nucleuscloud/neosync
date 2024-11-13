package enforcer

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	sqladapter "github.com/Blank-Xu/sql-adapter"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
)

func NewDefaultEnforcer(
	ctx context.Context,
	db *sql.DB,
	casbinTableName string,
) (casbin.IEnforcer, error) {
	adapter, err := newDefaultAdapter(ctx, db, casbinTableName)
	if err != nil {
		return nil, err
	}
	return newEnforcer(adapter)
}

func newEnforcer(
	adapter persist.Adapter,
) (casbin.IEnforcer, error) {
	m, err := model.NewModelFromString(neosyncRbacModel)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize casbin model from string: %w", err)
	}

	enforcer, err := casbin.NewSyncedCachedEnforcer(m, adapter)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize casbin synced cached enforcer: %w", err)
	}
	enforcer.EnableAutoSave(true) // seems to do this automatically but it doesn't hurt
	enforcer.SetExpireTime(30 * time.Second)
	return enforcer, nil
}

func newDefaultAdapter(
	ctx context.Context,
	db *sql.DB,
	tableName string,
) (persist.Adapter, error) {
	adapter, err := sqladapter.NewAdapterWithContext(ctx, db, "postgres", tableName)
	if err != nil {
		return nil, fmt.Errorf("unable to create casbin sql adapter: %w", err)
	}
	return adapter, nil
}
