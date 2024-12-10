package rbac

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
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
	return uUIDStrings(resp), nil
}

func uUIDStrings(values []pgtype.UUID) []string {
	outputs := make([]string, len(values))
	for idx := range values {
		outputs[idx] = uUIDString(values[idx])
	}
	return outputs
}
func uUIDString(value pgtype.UUID) string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", value.Bytes[0:4], value.Bytes[4:6], value.Bytes[6:8], value.Bytes[8:10], value.Bytes[10:16])
}
