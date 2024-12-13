package rbac

import (
	"context"

	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
)

type RbacDb struct {
	q    db_queries.Querier
	dbtx db_queries.DBTX
}

var _ Db = (*RbacDb)(nil)

func NewRbacDb(q db_queries.Querier, dbtx db_queries.DBTX) *RbacDb {
	return &RbacDb{q: q, dbtx: dbtx}
}

func (r *RbacDb) GetAccountIds(ctx context.Context) ([]string, error) {
	resp, err := r.q.GetAccountIds(ctx, r.dbtx)
	if err != nil {
		return nil, err
	}
	return neosyncdb.UUIDStrings(resp), nil
}
