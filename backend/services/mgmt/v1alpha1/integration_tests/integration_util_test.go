package integrationtests_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
)

func setAccountCreatedAt(
	ctx context.Context,
	querier db_queries.Querier,
	dbconn *pgxpool.Pool,
	accountId string,
	createdAt time.Time,
) error {
	accountUuid, err := neosyncdb.ToUuid(accountId)
	if err != nil {
		return err
	}
	_, err = querier.SetAccountCreatedAt(ctx, dbconn, db_queries.SetAccountCreatedAtParams{
		CreatedAt: pgtype.Timestamp{Time: createdAt, Valid: true},
		AccountId: accountUuid,
	})
	return err
}

func getAccountIds(t testing.TB, accounts []*mgmtv1alpha1.UserAccount) []string {
	t.Helper()
	output := []string{}
	for _, acc := range accounts {
		output = append(output, acc.GetId())
	}
	return output
}
