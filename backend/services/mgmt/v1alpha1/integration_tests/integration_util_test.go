package integrationtests_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
	tcneosyncapi "github.com/nucleuscloud/neosync/backend/pkg/integration-test"
)

func (s *IntegrationTestSuite) createPersonalAccount(
	ctx context.Context,
	userclient mgmtv1alpha1connect.UserAccountServiceClient,
) string {
	s.T().Helper()
	return tcneosyncapi.CreatePersonalAccount(ctx, s.T(), userclient)
}

func (s *IntegrationTestSuite) setAccountCreatedAt(
	ctx context.Context,
	accountId string,
	createdAt time.Time,
) error {
	accountUuid, err := neosyncdb.ToUuid(accountId)
	if err != nil {
		return err
	}
	_, err = s.NeosyncQuerier.SetAccountCreatedAt(ctx, s.Pgcontainer.DB, db_queries.SetAccountCreatedAtParams{
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
