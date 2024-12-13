package rbac

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/stdlib"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/internal/ee/rbac/enforcer"
	neomigrate "github.com/nucleuscloud/neosync/internal/migrate"
	"github.com/nucleuscloud/neosync/internal/testutil"
	testcontainers_postgres "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/postgres"
	"github.com/stretchr/testify/require"
)

func TestRbac(t *testing.T) {
	t.Parallel()
	ok := testutil.ShouldRunIntegrationTest()
	if !ok {
		return
	}

	ctx := context.Background()

	container, err := testcontainers_postgres.NewPostgresTestContainer(ctx)
	require.NoError(t, err)

	err = neomigrate.Up(ctx, container.URL, "../../../sql/postgresql/schema", testutil.GetTestLogger(t))
	require.NoError(t, err)

	rbacenforcer, err := enforcer.NewActiveEnforcer(ctx, stdlib.OpenDBFromPool(container.DB), "neosync_api.casbin_rule")
	require.NoError(t, err)

	rbacenforcer.EnableAutoSave(true)
	err = rbacenforcer.LoadPolicy()
	require.NoError(t, err)

	rbacclient := New(rbacenforcer)

	t.Run("account_admin", func(t *testing.T) {
		t.Parallel()

		accountId := uuid.NewString()
		userId := uuid.NewString()
		err := rbacclient.SetupNewAccount(ctx, accountId, testutil.GetTestLogger(t))
		require.NoError(t, err)

		// Set user as account admin
		err = rbacclient.SetAccountRole(ctx, NewUserIdEntity(userId), NewAccountIdEntity(accountId), mgmtv1alpha1.AccountRole_ACCOUNT_ROLE_ADMIN)
		require.NoError(t, err)

		// Test Account permissions
		err = rbacclient.EnforceAccount(ctx, NewUserIdEntity(userId), NewAccountIdEntity(accountId), AccountAction_View)
		require.NoError(t, err)

		err = rbacclient.EnforceAccount(ctx, NewUserIdEntity(userId), NewAccountIdEntity(accountId), AccountAction_Edit)
		require.NoError(t, err)

		// Test Job permissions
		jobId := uuid.NewString()
		err = rbacclient.EnforceJob(ctx, NewUserIdEntity(userId), NewAccountIdEntity(accountId), NewJobIdEntity(jobId), JobAction_Create)
		require.NoError(t, err)

		// Test Connection permissions
		connectionId := uuid.NewString()
		err = rbacclient.EnforceConnection(ctx, NewUserIdEntity(userId), NewAccountIdEntity(accountId), NewConnectionIdEntity(connectionId), ConnectionAction_Create)
		require.NoError(t, err)
	})

	t.Run("job_developer", func(t *testing.T) {
		t.Parallel()

		accountId := uuid.NewString()
		userId := uuid.NewString()
		err := rbacclient.SetupNewAccount(ctx, accountId, testutil.GetTestLogger(t))
		require.NoError(t, err)

		// Set user as job developer
		err = rbacclient.SetAccountRole(ctx, NewUserIdEntity(userId), NewAccountIdEntity(accountId), mgmtv1alpha1.AccountRole_ACCOUNT_ROLE_JOB_DEVELOPER)
		require.NoError(t, err)

		// Test Account permissions (should only have view)
		err = rbacclient.EnforceAccount(ctx, NewUserIdEntity(userId), NewAccountIdEntity(accountId), AccountAction_View)
		require.NoError(t, err)

		err = rbacclient.EnforceAccount(ctx, NewUserIdEntity(userId), NewAccountIdEntity(accountId), AccountAction_Edit)
		require.Error(t, err)
		require.ErrorContains(t, err, "user does not have permission to edit account")

		// Test Job permissions (should have all)
		jobId := uuid.NewString()
		err = rbacclient.EnforceJob(ctx, NewUserIdEntity(userId), NewAccountIdEntity(accountId), NewJobIdEntity(jobId), JobAction_Create)
		require.NoError(t, err)
	})

	t.Run("job_viewer", func(t *testing.T) {
		t.Parallel()

		accountId := uuid.NewString()
		userId := uuid.NewString()
		err := rbacclient.SetupNewAccount(ctx, accountId, testutil.GetTestLogger(t))
		require.NoError(t, err)

		// Set user as job viewer
		err = rbacclient.SetAccountRole(ctx, NewUserIdEntity(userId), NewAccountIdEntity(accountId), mgmtv1alpha1.AccountRole_ACCOUNT_ROLE_JOB_VIEWER)
		require.NoError(t, err)

		// Test Account permissions (should only have view)
		err = rbacclient.EnforceAccount(ctx, NewUserIdEntity(userId), NewAccountIdEntity(accountId), AccountAction_View)
		require.NoError(t, err)

		// Test Job permissions (should only have view and execute)
		jobId := uuid.NewString()
		err = rbacclient.EnforceJob(ctx, NewUserIdEntity(userId), NewAccountIdEntity(accountId), NewJobIdEntity(jobId), JobAction_View)
		require.NoError(t, err)

		err = rbacclient.EnforceJob(ctx, NewUserIdEntity(userId), NewAccountIdEntity(accountId), NewJobIdEntity(jobId), JobAction_Execute)
		require.NoError(t, err)

		err = rbacclient.EnforceJob(ctx, NewUserIdEntity(userId), NewAccountIdEntity(accountId), NewJobIdEntity(jobId), JobAction_Create)
		require.Error(t, err)
		require.ErrorContains(t, err, "user does not have permission to create job")
	})

	t.Run("role_changes", func(t *testing.T) {
		t.Parallel()

		accountId := uuid.NewString()
		userId := uuid.NewString()
		err := rbacclient.SetupNewAccount(ctx, accountId, testutil.GetTestLogger(t))
		require.NoError(t, err)

		// Set and remove roles
		err = rbacclient.SetAccountRole(ctx, NewUserIdEntity(userId), NewAccountIdEntity(accountId), mgmtv1alpha1.AccountRole_ACCOUNT_ROLE_ADMIN)
		require.NoError(t, err)

		err = rbacclient.RemoveAccountRole(ctx, NewUserIdEntity(userId), NewAccountIdEntity(accountId), mgmtv1alpha1.AccountRole_ACCOUNT_ROLE_ADMIN)
		require.NoError(t, err)

		// Verify permissions are removed
		err = rbacclient.EnforceAccount(ctx, NewUserIdEntity(userId), NewAccountIdEntity(accountId), AccountAction_View)
		require.Error(t, err)
		require.ErrorContains(t, err, "user does not have permission to view account")
	})
}
