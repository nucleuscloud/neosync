package neosyncdb

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	neomigrate "github.com/nucleuscloud/neosync/internal/migrate"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	testpg "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/sync/errgroup"
)

var (
	discardLogger = slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
)

type IntegrationTestSuite struct {
	suite.Suite

	pgpool *pgxpool.Pool

	ctx context.Context

	pgcontainer   *testpg.PostgresContainer
	connstr       string
	migrationsDir string

	db *NeosyncDb
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()

	pgcontainer, err := testpg.Run(
		s.ctx,
		"postgres:15",
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second),
		),
	)
	if err != nil {
		panic(err)
	}
	s.pgcontainer = pgcontainer
	connstr, err := pgcontainer.ConnectionString(s.ctx, "sslmode=disable")
	if err != nil {
		panic(err)
	}
	s.connstr = connstr

	pool, err := pgxpool.New(s.ctx, connstr)
	if err != nil {
		panic(err)
	}
	s.pgpool = pool
	s.migrationsDir = "../../sql/postgresql/schema"

	s.db = New(pool, db_queries.New())
}

// Runs before each test
func (s *IntegrationTestSuite) SetupTest() {
	discardLogger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	err := neomigrate.Up(s.ctx, s.connstr, s.migrationsDir, discardLogger)
	if err != nil {
		panic(err)
	}
}

func (s *IntegrationTestSuite) TearDownTest() {
	// Dropping here because 1) more efficient and 2) we have a bad down migration
	// _jobs-connection-id-null.down that breaks due to having a null connection_id column.
	// we should do something about that at some point. Running this single drop is easier though
	_, err := s.pgpool.Exec(s.ctx, "DROP SCHEMA IF EXISTS neosync_api CASCADE")
	if err != nil {
		panic(err)
	}
	_, err = s.pgpool.Exec(s.ctx, "DROP TABLE IF EXISTS public.schema_migrations")
	if err != nil {
		panic(err)
	}
}

func (s *IntegrationTestSuite) TearDownSuite() {
	if s.pgpool != nil {
		s.pgpool.Close()
	}
	if s.pgcontainer != nil {
		err := s.pgcontainer.Terminate(s.ctx)
		if err != nil {
			panic(err)
		}
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	evkey := "INTEGRATION_TESTS_ENABLED"
	shouldRun := os.Getenv(evkey)
	if shouldRun != "1" {
		slog.Warn(fmt.Sprintf("skipping integration tests, set %s=1 to enable", evkey))
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

		account, err := s.db.CreateTeamAccount(s.ctx, user.ID, "myteam", discardLogger)
		requireNoErrResp(t, account, err)
	})

	t.Run("already exists", func(t *testing.T) {
		user := s.setUser(t, s.ctx, "foo1")

		account, err := s.db.CreateTeamAccount(s.ctx, user.ID, "myteam", discardLogger)
		requireNoErrResp(t, account, err)

		account1, err := s.db.CreateTeamAccount(s.ctx, user.ID, "myteam", discardLogger)
		requireErrResp(t, account1, err)
		alreadyExists := nucleuserrors.NewAlreadyExists("")
		require.ErrorAs(t, err, &alreadyExists)
	})
}

func (s *IntegrationTestSuite) Test_UpsertStripeCustomerId() {
	t := s.T()

	user := s.setUser(t, s.ctx, "foo")

	t.Run("new customer id", func(t *testing.T) {
		account, err := s.db.CreateTeamAccount(s.ctx, user.ID, "myteam", discardLogger)
		requireNoErrResp(t, account, err)
		require.False(t, account.StripeCustomerID.Valid)

		account, err = s.db.UpsertStripeCustomerId(s.ctx, account.ID, func(ctx context.Context, account db_queries.NeosyncApiAccount) (string, error) {
			return "testid", nil
		}, discardLogger)
		requireNoErrResp(t, account, err)
		require.True(t, account.StripeCustomerID.Valid)
		require.Equal(t, "testid", account.StripeCustomerID.String)
	})

	t.Run("only first one", func(t *testing.T) {
		account, err := s.db.CreateTeamAccount(s.ctx, user.ID, "myteam1", discardLogger)
		requireNoErrResp(t, account, err)
		require.False(t, account.StripeCustomerID.Valid)

		firstid := "testid"
		account, err = s.db.UpsertStripeCustomerId(s.ctx, account.ID, func(ctx context.Context, account db_queries.NeosyncApiAccount) (string, error) {
			return firstid, nil
		}, discardLogger)
		requireNoErrResp(t, account, err)
		require.True(t, account.StripeCustomerID.Valid)
		require.Equal(t, firstid, account.StripeCustomerID.String)

		account, err = s.db.UpsertStripeCustomerId(s.ctx, account.ID, func(ctx context.Context, account db_queries.NeosyncApiAccount) (string, error) {
			return "secondid", nil
		}, discardLogger)
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
		}, discardLogger)
		requireErrResp(t, account, err)
	})
}

func (s *IntegrationTestSuite) Test_CreateTeamAccountInvite() {
	t := s.T()

	user := s.setUser(t, s.ctx, "foo")
	account, err := s.db.CreateTeamAccount(s.ctx, user.ID, "myteam1", discardLogger)
	requireNoErrResp(t, account, err)

	t.Run("new invite", func(t *testing.T) {
		invite, err := s.db.CreateTeamAccountInvite(s.ctx, account.ID, user.ID, "foo2@example.com", getFutureTs(t, 1*time.Hour))
		requireNoErrResp(t, invite, err)
	})

	t.Run("expire old invites", func(t *testing.T) {
		invite, err := s.db.CreateTeamAccountInvite(s.ctx, account.ID, user.ID, "foo2@example.com", getFutureTs(t, 1*time.Hour))
		requireNoErrResp(t, invite, err)

		invite2, err := s.db.CreateTeamAccountInvite(s.ctx, account.ID, user.ID, "foo@example.com", getFutureTs(t, 1*time.Hour))
		requireNoErrResp(t, invite2, err)
		now := time.Now()

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
	account, err := s.db.CreateTeamAccount(s.ctx, user.ID, "myteam1", discardLogger)
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
	account, err := s.db.CreateTeamAccount(s.ctx, user.ID, "myteam1", discardLogger)
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
