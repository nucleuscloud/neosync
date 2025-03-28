package enforcer

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	sqladapter "github.com/nucleuscloud/sql-adapter"
)

// The default casbin enforcer with a SQL-enabled backend
func NewActiveEnforcer(
	ctx context.Context,
	db *sql.DB,
	casbinTableName string,
) (casbin.IEnforcer, error) {
	adapter, err := newSqlAdapter(ctx, db, casbinTableName)
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

	enforcer, err := casbin.NewSyncedEnforcer(m, adapter)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize casbin synced cached enforcer: %w", err)
	}
	enforcer.EnableAutoSave(
		true,
	) // seems to do this automatically but it doesn't hurt
	enforcer.StartAutoLoadPolicy(
		time.Second * 10,
	) // allows HA between neosync-api instances or backend changes to RBAC policies to be picked up.
	return enforcer, nil
}

func newSqlAdapter(
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
