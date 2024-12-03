package neosyncdb

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	pg_models "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"
	neomigrate "github.com/nucleuscloud/neosync/internal/migrate"
	"github.com/nucleuscloud/neosync/internal/testutil"
	tcpostgres "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/postgres"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/sync/errgroup"
)

type IntegrationTestSuite struct {
	suite.Suite

	ctx context.Context

	pgcontainer   *tcpostgres.PostgresTestContainer
	migrationsDir string

	db *NeosyncDb
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()

	pgcontainer, err := tcpostgres.NewPostgresTestContainer(s.ctx)
	if err != nil {
		panic(err)
	}
	s.pgcontainer = pgcontainer

	s.migrationsDir = "../../sql/postgresql/schema"

	s.db = New(s.pgcontainer.DB, db_queries.New())
}

// Runs before each test
func (s *IntegrationTestSuite) SetupTest() {
	err := neomigrate.Up(s.ctx, s.pgcontainer.URL, s.migrationsDir, testutil.GetTestLogger(s.T()))
	if err != nil {
		panic(err)
	}
}

func (s *IntegrationTestSuite) TearDownTest() {
	// Dropping here because 1) more efficient and 2) we have a bad down migration
	// _jobs-connection-id-null.down that breaks due to having a null connection_id column.
	// we should do something about that at some point. Running this single drop is easier though
	_, err := s.pgcontainer.DB.Exec(s.ctx, "DROP SCHEMA IF EXISTS neosync_api CASCADE")
	if err != nil {
		panic(err)
	}
	_, err = s.pgcontainer.DB.Exec(s.ctx, "DROP TABLE IF EXISTS public.schema_migrations")
	if err != nil {
		panic(err)
	}
}

func (s *IntegrationTestSuite) TearDownSuite() {
	if s.pgcontainer != nil {
		err := s.pgcontainer.TearDown(s.ctx)
		if err != nil {
			panic(err)
		}
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	ok := testutil.ShouldRunIntegrationTest()
	if !ok {
		return
	}
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) Test_SetUserByAuth0Id() {
	t := s.T()

	t.Run("new user", func(t *testing.T) {
		resp, err := s.db.SetUserByAuthSub(s.ctx, "foo")
		requireNoErrResp(t, resp, err)
		require.NotNil(t, resp.ID)
	})

	t.Run("idempotent", func(t *testing.T) {
		resp, err := s.db.SetUserByAuthSub(s.ctx, "myid")
		requireNoErrResp(t, resp, err)

		resp2, err := s.db.SetUserByAuthSub(s.ctx, "myid")
		requireNoErrResp(t, resp2, err)

		uid1 := UUIDString(resp.ID)
		uid2 := UUIDString(resp2.ID)
		require.Equal(t, uid1, uid2)
	})
}

func (s *IntegrationTestSuite) setUser(t testing.TB, ctx context.Context, sub string) *db_queries.NeosyncApiUser {
	resp, err := s.db.SetUserByAuthSub(ctx, sub)
	requireNoErrResp(t, resp, err)
	return resp
}

func (s *IntegrationTestSuite) Test_SetPersonalAccount() {
	t := s.T()

	t.Run("new account", func(t *testing.T) {
		user := s.setUser(t, s.ctx, "foo")
		maxAllowed := int64(100)
		resp, err := s.db.SetPersonalAccount(s.ctx, user.ID, &maxAllowed)
		requireNoErrResp(t, resp, err)
	})

	t.Run("idempotent", func(t *testing.T) {
		user := s.setUser(t, s.ctx, "foo1")
		maxAllowed := int64(100)

		resp, err := s.db.SetPersonalAccount(s.ctx, user.ID, &maxAllowed)
		requireNoErrResp(t, resp, err)

		resp2, err := s.db.SetPersonalAccount(s.ctx, user.ID, &maxAllowed)
		requireNoErrResp(t, resp2, err)

		uid1 := UUIDString(resp.ID)
		uid2 := UUIDString(resp2.ID)
		require.Equal(t, uid1, uid2)
	})

	t.Run("idempotent - parallel", func(t *testing.T) {
		user := s.setUser(t, s.ctx, "foo2")
		maxAllowed := int64(100)

		errgrp, errctx := errgroup.WithContext(s.ctx)

		var uid1 string
		errgrp.Go(func() error {
			resp, err := s.db.SetPersonalAccount(errctx, user.ID, &maxAllowed)
			assertNoErrResp(t, resp, err)
			uid1 = UUIDString(resp.ID)
			return nil
		})

		var uid2 string
		errgrp.Go(func() error {
			resp, err := s.db.SetPersonalAccount(errctx, user.ID, &maxAllowed)
			assertNoErrResp(t, resp, err)
			uid2 = UUIDString(resp.ID)
			return nil
		})

		err := errgrp.Wait()
		require.NoError(t, err)
		require.Equal(t, uid1, uid2)
	})
}

func (s *IntegrationTestSuite) Test_CreateTeamAccount() {
	t := s.T()

	t.Run("new account", func(t *testing.T) {
		user := s.setUser(t, s.ctx, "foo")

		account, err := s.db.CreateTeamAccount(s.ctx, user.ID, "myteam", testutil.GetTestLogger(t))
		requireNoErrResp(t, account, err)
	})

	t.Run("already exists", func(t *testing.T) {
		user := s.setUser(t, s.ctx, "foo1")

		account, err := s.db.CreateTeamAccount(s.ctx, user.ID, "myteam", testutil.GetTestLogger(t))
		requireNoErrResp(t, account, err)

		account1, err := s.db.CreateTeamAccount(s.ctx, user.ID, "myteam", testutil.GetTestLogger(t))
		requireErrResp(t, account1, err)
		alreadyExists := nucleuserrors.NewAlreadyExists("")
		require.ErrorAs(t, err, &alreadyExists)
	})
}

func (s *IntegrationTestSuite) Test_ConvertPersonalToTeamAccount() {
	t := s.T()

	t.Run("success", func(t *testing.T) {
		user := s.setUser(t, s.ctx, "convertPtoT1")
		maxAllowedRecords := int64(100)
		account, err := s.db.SetPersonalAccount(s.ctx, user.ID, &maxAllowedRecords)
		requireNoErrResp(t, account, err)

		newTeamName := "newteam"
		resp, err := s.db.ConvertPersonalToTeamAccount(s.ctx, &ConvertPersonalToTeamAccountRequest{
			UserId:            user.ID,
			PersonalAccountId: account.ID,
			TeamName:          newTeamName,
		}, testutil.GetTestLogger(t))
		requireNoErrResp(t, resp, err)

		require.Equal(t, UUIDString(account.ID), UUIDString(resp.TeamAccount.ID), "the new team account must be the same id as the old account")
		require.Equal(t, AccountType_Team, AccountType(resp.TeamAccount.AccountType))
		require.Equal(t, newTeamName, resp.TeamAccount.AccountSlug)
		require.NotEqual(t, UUIDString(account.ID), UUIDString(resp.PersonalAccount.ID), "the new personal account must not have the same id as the old one")
		require.False(t, resp.TeamAccount.MaxAllowedRecords.Valid, "team account must not have any max allowed records set")
		require.Equal(t, maxAllowedRecords, resp.PersonalAccount.MaxAllowedRecords.Int64, "max allowed records must persist on new personal account")
	})

	t.Run("invalid account type", func(t *testing.T) {
		user := s.setUser(t, s.ctx, "convertPtoT2")
		account, err := s.db.CreateTeamAccount(s.ctx, user.ID, "myteam", testutil.GetTestLogger(t))
		requireNoErrResp(t, account, err)

		resp, err := s.db.ConvertPersonalToTeamAccount(s.ctx, &ConvertPersonalToTeamAccountRequest{UserId: user.ID, PersonalAccountId: account.ID, TeamName: "myteam2"}, testutil.GetTestLogger(t))
		requireErrResp(t, resp, err)
		badreqerror := nucleuserrors.NewBadRequest("")
		require.ErrorAs(t, err, &badreqerror)
	})
}

func (s *IntegrationTestSuite) Test_UpsertStripeCustomerId() {
	t := s.T()

	user := s.setUser(t, s.ctx, "foo")

	t.Run("new customer id", func(t *testing.T) {
		account, err := s.db.CreateTeamAccount(s.ctx, user.ID, "myteam", testutil.GetTestLogger(t))
		requireNoErrResp(t, account, err)
		require.False(t, account.StripeCustomerID.Valid)

		account, err = s.db.UpsertStripeCustomerId(s.ctx, account.ID, func(ctx context.Context, account db_queries.NeosyncApiAccount) (string, error) {
			return "testid", nil
		}, testutil.GetTestLogger(t))
		requireNoErrResp(t, account, err)
		require.True(t, account.StripeCustomerID.Valid)
		require.Equal(t, "testid", account.StripeCustomerID.String)
	})

	t.Run("only first one", func(t *testing.T) {
		account, err := s.db.CreateTeamAccount(s.ctx, user.ID, "myteam1", testutil.GetTestLogger(t))
		requireNoErrResp(t, account, err)
		require.False(t, account.StripeCustomerID.Valid)

		firstid := "testid"
		account, err = s.db.UpsertStripeCustomerId(s.ctx, account.ID, func(ctx context.Context, account db_queries.NeosyncApiAccount) (string, error) {
			return firstid, nil
		}, testutil.GetTestLogger(t))
		requireNoErrResp(t, account, err)
		require.True(t, account.StripeCustomerID.Valid)
		require.Equal(t, firstid, account.StripeCustomerID.String)

		account, err = s.db.UpsertStripeCustomerId(s.ctx, account.ID, func(ctx context.Context, account db_queries.NeosyncApiAccount) (string, error) {
			return "secondid", nil
		}, testutil.GetTestLogger(t))
		requireNoErrResp(t, account, err)
		require.True(t, account.StripeCustomerID.Valid)
		require.Equal(t, firstid, account.StripeCustomerID.String)
	})

	t.Run("personal not allowed", func(t *testing.T) {
		account, err := s.db.SetPersonalAccount(s.ctx, user.ID, nil)
		requireNoErrResp(t, account, err)
		require.False(t, account.StripeCustomerID.Valid)

		account, err = s.db.UpsertStripeCustomerId(s.ctx, account.ID, func(ctx context.Context, account db_queries.NeosyncApiAccount) (string, error) {
			return "testid", nil
		}, testutil.GetTestLogger(t))
		requireErrResp(t, account, err)
	})
}

func (s *IntegrationTestSuite) Test_CreateTeamAccountInvite() {
	t := s.T()

	user := s.setUser(t, s.ctx, "foo")
	account, err := s.db.CreateTeamAccount(s.ctx, user.ID, "myteam1", testutil.GetTestLogger(t))
	requireNoErrResp(t, account, err)

	t.Run("new invite", func(t *testing.T) {
		invite, err := s.db.CreateTeamAccountInvite(s.ctx, account.ID, user.ID, "foo2@example.com", getFutureTs(t, 1*time.Hour))
		requireNoErrResp(t, invite, err)
	})

	t.Run("expire old invites", func(t *testing.T) {
		invite, err := s.db.CreateTeamAccountInvite(s.ctx, account.ID, user.ID, "foo2@example.com", getFutureTs(t, 48*time.Hour))
		requireNoErrResp(t, invite, err)

		invite2, err := s.db.CreateTeamAccountInvite(s.ctx, account.ID, user.ID, "foo2@example.com", getFutureTs(t, 48*time.Hour))
		requireNoErrResp(t, invite2, err)
		// Add time here as the expired invites as updated to CURRENT_TIMESTAMP, so this reduces flakiness
		now := time.Now().Add(5 * time.Second)

		oldinvite1, err := s.db.Q.GetAccountInvite(s.ctx, s.db.Db, invite.ID)
		requireNoErrResp(t, oldinvite1, err)
		require.Greater(t, now.Unix(), oldinvite1.ExpiresAt.Time.Unix())
	})

	t.Run("personal not allowed", func(t *testing.T) {
		account, err := s.db.SetPersonalAccount(s.ctx, user.ID, nil)
		requireNoErrResp(t, account, err)

		invite, err := s.db.CreateTeamAccountInvite(s.ctx, account.ID, user.ID, "foo@example.com", getFutureTs(t, 1*time.Hour))
		requireErrResp(t, invite, err)
		forbiddin := nucleuserrors.NewForbidden("")
		require.ErrorAs(t, err, &forbiddin)
	})
}

func (s *IntegrationTestSuite) Test_ValidateInviteAddUserToAccount() {
	t := s.T()

	user := s.setUser(t, s.ctx, "foo")
	account, err := s.db.CreateTeamAccount(s.ctx, user.ID, "myteam1", testutil.GetTestLogger(t))
	requireNoErrResp(t, account, err)

	t.Run("accept invite", func(t *testing.T) {
		invite, err := s.db.CreateTeamAccountInvite(s.ctx, account.ID, user.ID, "foo2@example.com", getFutureTs(t, 24*time.Hour))
		requireNoErrResp(t, invite, err)

		user2 := s.setUser(t, s.ctx, "foo2")

		accountId, err := s.db.ValidateInviteAddUserToAccount(s.ctx, user2.ID, invite.Token, "foo2@example.com")
		requireNoErrResp(t, accountId, err)
	})

	t.Run("expired invite", func(t *testing.T) {
		invite, err := s.db.CreateTeamAccountInvite(s.ctx, account.ID, user.ID, "foo3@example.com", getFutureTs(t, -1*time.Hour))
		requireNoErrResp(t, invite, err)

		user3 := s.setUser(t, s.ctx, "foo3")

		accountId, err := s.db.ValidateInviteAddUserToAccount(s.ctx, user3.ID, invite.Token, "foo3@example.com")
		require.Error(t, err)
		require.False(t, accountId.Valid)
		forbidden := nucleuserrors.NewForbidden("")
		require.ErrorAs(t, err, &forbidden)
	})

	t.Run("incorrect email", func(t *testing.T) {
		invite, err := s.db.CreateTeamAccountInvite(s.ctx, account.ID, user.ID, "foo4@example.com", getFutureTs(t, -1*time.Hour))
		requireNoErrResp(t, invite, err)

		user4 := s.setUser(t, s.ctx, "foo3")

		accountId, err := s.db.ValidateInviteAddUserToAccount(s.ctx, user4.ID, invite.Token, "blah@example.com")
		require.Error(t, err)
		require.False(t, accountId.Valid)
		badrequest := nucleuserrors.NewBadRequest("")
		require.ErrorAs(t, err, &badrequest)
		t.Log(err.Error())
	})
}

func (s *IntegrationTestSuite) Test_CreateAccountApiKey() {
	t := s.T()

	user := s.setUser(t, s.ctx, "foo")
	account, err := s.db.CreateTeamAccount(s.ctx, user.ID, "myteam1", testutil.GetTestLogger(t))
	requireNoErrResp(t, account, err)

	key, err := s.db.CreateAccountApikey(s.ctx, &CreateAccountApiKeyRequest{
		KeyName:           "foo",
		KeyValue:          "bar",
		AccountUuid:       account.ID,
		CreatedByUserUuid: user.ID,
		ExpiresAt:         getFutureTs(t, 24*time.Hour),
	})
	requireNoErrResp(t, key, err)
}

func (s *IntegrationTestSuite) Test_CreateJob() {
	t := s.T()

	user := s.setUser(t, s.ctx, "foo")
	account, err := s.db.CreateTeamAccount(s.ctx, user.ID, "myteam1", testutil.GetTestLogger(t))
	requireNoErrResp(t, account, err)

	connection, err := s.db.Q.CreateConnection(s.ctx, s.db.Db, db_queries.CreateConnectionParams{
		Name:             "foo",
		AccountID:        account.ID,
		ConnectionConfig: &pg_models.ConnectionConfig{},
		CreatedByID:      user.ID,
		UpdatedByID:      user.ID,
	})
	requireNoErrResp(t, connection, err)

	job, err := s.db.CreateJob(s.ctx, &db_queries.CreateJobParams{
		Name:               "foo",
		AccountID:          account.ID,
		Status:             1,
		ConnectionOptions:  &pg_models.JobSourceOptions{PostgresOptions: &pg_models.PostgresSourceOptions{HaltOnNewColumnAddition: true}},
		Mappings:           []*pg_models.JobMapping{{Schema: "foo", Table: "bar", Column: "baz"}},
		CronSchedule:       pgtype.Text{String: "blah", Valid: true},
		CreatedByID:        user.ID,
		UpdatedByID:        user.ID,
		WorkflowOptions:    &pg_models.WorkflowOptions{},
		SyncOptions:        &pg_models.ActivityOptions{},
		VirtualForeignKeys: []*pg_models.VirtualForeignConstraint{},
	}, []*CreateJobConnectionDestination{
		{
			ConnectionId: connection.ID,
			Options:      &pg_models.JobDestinationOptions{},
		},
	})
	requireNoErrResp(t, job, err)
}

func (s *IntegrationTestSuite) Test_SetSourceSubsets() {
	t := s.T()

	user := s.setUser(t, s.ctx, "foo")
	account, err := s.db.CreateTeamAccount(s.ctx, user.ID, "myteam1", testutil.GetTestLogger(t))
	requireNoErrResp(t, account, err)

	job, err := s.db.CreateJob(s.ctx, &db_queries.CreateJobParams{
		Name:               "foo",
		AccountID:          account.ID,
		Status:             1,
		ConnectionOptions:  &pg_models.JobSourceOptions{PostgresOptions: &pg_models.PostgresSourceOptions{HaltOnNewColumnAddition: true}},
		Mappings:           []*pg_models.JobMapping{{Schema: "foo", Table: "bar", Column: "baz"}},
		CronSchedule:       pgtype.Text{String: "blah", Valid: true},
		CreatedByID:        user.ID,
		UpdatedByID:        user.ID,
		WorkflowOptions:    &pg_models.WorkflowOptions{},
		SyncOptions:        &pg_models.ActivityOptions{},
		VirtualForeignKeys: []*pg_models.VirtualForeignConstraint{},
	}, []*CreateJobConnectionDestination{})
	requireNoErrResp(t, job, err)

	where := "blah"

	t.Run("postgres", func(t *testing.T) {
		err := s.db.SetSourceSubsets(s.ctx, job.ID, &mgmtv1alpha1.JobSourceSqlSubetSchemas{
			Schemas: &mgmtv1alpha1.JobSourceSqlSubetSchemas_PostgresSubset{
				PostgresSubset: &mgmtv1alpha1.PostgresSourceSchemaSubset{
					PostgresSchemas: []*mgmtv1alpha1.PostgresSourceSchemaOption{
						{Schema: "foo", Tables: []*mgmtv1alpha1.PostgresSourceTableOption{{Table: "foo", WhereClause: &where}}},
					},
				},
			},
		}, false, user.ID)
		require.NoError(t, err)
	})

	t.Run("mysql", func(t *testing.T) {
		err := s.db.SetSourceSubsets(s.ctx, job.ID, &mgmtv1alpha1.JobSourceSqlSubetSchemas{
			Schemas: &mgmtv1alpha1.JobSourceSqlSubetSchemas_MysqlSubset{
				MysqlSubset: &mgmtv1alpha1.MysqlSourceSchemaSubset{
					MysqlSchemas: []*mgmtv1alpha1.MysqlSourceSchemaOption{
						{Schema: "foo", Tables: []*mgmtv1alpha1.MysqlSourceTableOption{{Table: "foo", WhereClause: &where}}},
					},
				},
			},
		}, false, user.ID)
		require.NoError(t, err)
	})

	t.Run("mssql", func(t *testing.T) {
		err := s.db.SetSourceSubsets(s.ctx, job.ID, &mgmtv1alpha1.JobSourceSqlSubetSchemas{
			Schemas: &mgmtv1alpha1.JobSourceSqlSubetSchemas_MssqlSubset{
				MssqlSubset: &mgmtv1alpha1.MssqlSourceSchemaSubset{
					MssqlSchemas: []*mgmtv1alpha1.MssqlSourceSchemaOption{
						{Schema: "foo", Tables: []*mgmtv1alpha1.MssqlSourceTableOption{{Table: "foo", WhereClause: &where}}},
					},
				},
			},
		}, false, user.ID)
		require.NoError(t, err)
	})

	t.Run("dynamodb", func(t *testing.T) {
		err := s.db.SetSourceSubsets(s.ctx, job.ID, &mgmtv1alpha1.JobSourceSqlSubetSchemas{
			Schemas: &mgmtv1alpha1.JobSourceSqlSubetSchemas_DynamodbSubset{
				DynamodbSubset: &mgmtv1alpha1.DynamoDBSourceSchemaSubset{
					Tables: []*mgmtv1alpha1.DynamoDBSourceTableOption{
						{Table: "foo", WhereClause: &where},
					},
				},
			},
		}, false, user.ID)
		require.NoError(t, err)
	})
}
