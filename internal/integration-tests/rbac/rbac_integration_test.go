package rbac

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/stdlib"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/internal/ee/rbac"
	"github.com/nucleuscloud/neosync/internal/ee/rbac/enforcer"
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

	err = neomigrate.Up(ctx, container.URL, "../../../backend/sql/postgresql/schema", testutil.GetTestLogger(t))
	require.NoError(t, err)

	rbacenforcer, err := enforcer.NewActiveEnforcer(ctx, stdlib.OpenDBFromPool(container.DB), "neosync_api.casbin_rule")
	require.NoError(t, err)

	rbacenforcer.EnableAutoSave(true)
	err = rbacenforcer.LoadPolicy()
	require.NoError(t, err)

	rbacclient := rbac.New(rbacenforcer)

	t.Run("account_admin", func(t *testing.T) {
		t.Parallel()

		accountId := uuid.NewString()
		userId := uuid.NewString()
		err := rbacclient.SetupNewAccount(ctx, accountId, testutil.GetTestLogger(t))
		require.NoError(t, err)

		// Set user as account admin
		err = rbacclient.SetAccountRole(ctx, rbac.NewUserIdEntity(userId), rbac.NewAccountIdEntity(accountId), mgmtv1alpha1.AccountRole_ACCOUNT_ROLE_ADMIN)
		require.NoError(t, err)

		// Test Account permissions
		err = rbacclient.EnforceAccount(ctx, rbac.NewUserIdEntity(userId), rbac.NewAccountIdEntity(accountId), rbac.AccountAction_View)
		require.NoError(t, err)

		err = rbacclient.EnforceAccount(ctx, rbac.NewUserIdEntity(userId), rbac.NewAccountIdEntity(accountId), rbac.AccountAction_Edit)
		require.NoError(t, err)

		// Test Job permissions
		jobId := uuid.NewString()
		err = rbacclient.EnforceJob(ctx, rbac.NewUserIdEntity(userId), rbac.NewAccountIdEntity(accountId), rbac.NewJobIdEntity(jobId), rbac.JobAction_Create)
		require.NoError(t, err)

		// Test Connection permissions
		connectionId := uuid.NewString()
		err = rbacclient.EnforceConnection(ctx, rbac.NewUserIdEntity(userId), rbac.NewAccountIdEntity(accountId), rbac.NewConnectionIdEntity(connectionId), rbac.ConnectionAction_Create)
		require.NoError(t, err)
	})

	t.Run("job_developer", func(t *testing.T) {
		t.Parallel()

		accountId := uuid.NewString()
		userId := uuid.NewString()
		err := rbacclient.SetupNewAccount(ctx, accountId, testutil.GetTestLogger(t))
		require.NoError(t, err)

		// Set user as job developer
		err = rbacclient.SetAccountRole(ctx, rbac.NewUserIdEntity(userId), rbac.NewAccountIdEntity(accountId), mgmtv1alpha1.AccountRole_ACCOUNT_ROLE_JOB_DEVELOPER)
		require.NoError(t, err)

		// Test Account permissions (should only have view)
		err = rbacclient.EnforceAccount(ctx, rbac.NewUserIdEntity(userId), rbac.NewAccountIdEntity(accountId), rbac.AccountAction_View)
		require.NoError(t, err)

		err = rbacclient.EnforceAccount(ctx, rbac.NewUserIdEntity(userId), rbac.NewAccountIdEntity(accountId), rbac.AccountAction_Edit)
		require.Error(t, err)
		require.ErrorContains(t, err, "user does not have permission to edit account")

		// Test Job permissions (should have all)
		jobId := uuid.NewString()
		err = rbacclient.EnforceJob(ctx, rbac.NewUserIdEntity(userId), rbac.NewAccountIdEntity(accountId), rbac.NewJobIdEntity(jobId), rbac.JobAction_Create)
		require.NoError(t, err)
	})

	t.Run("job_executor", func(t *testing.T) {
		t.Parallel()

		accountId := uuid.NewString()
		userId := uuid.NewString()
		err := rbacclient.SetupNewAccount(ctx, accountId, testutil.GetTestLogger(t))
		require.NoError(t, err)

		// Set user as job executor
		err = rbacclient.SetAccountRole(ctx, rbac.NewUserIdEntity(userId), rbac.NewAccountIdEntity(accountId), mgmtv1alpha1.AccountRole_ACCOUNT_ROLE_JOB_EXECUTOR)
		require.NoError(t, err)

		// Test Account permissions (should only have view)
		err = rbacclient.EnforceAccount(ctx, rbac.NewUserIdEntity(userId), rbac.NewAccountIdEntity(accountId), rbac.AccountAction_View)
		require.NoError(t, err)

		// Test Job permissions (should only have view and execute)
		jobId := uuid.NewString()
		err = rbacclient.EnforceJob(ctx, rbac.NewUserIdEntity(userId), rbac.NewAccountIdEntity(accountId), rbac.NewJobIdEntity(jobId), rbac.JobAction_View)
		require.NoError(t, err)

		err = rbacclient.EnforceJob(ctx, rbac.NewUserIdEntity(userId), rbac.NewAccountIdEntity(accountId), rbac.NewJobIdEntity(jobId), rbac.JobAction_Execute)
		require.NoError(t, err)

		err = rbacclient.EnforceJob(ctx, rbac.NewUserIdEntity(userId), rbac.NewAccountIdEntity(accountId), rbac.NewJobIdEntity(jobId), rbac.JobAction_Create)
		require.Error(t, err)
		require.ErrorContains(t, err, "user does not have permission to create job")
	})

	t.Run("job_viewer", func(t *testing.T) {
		t.Parallel()

		accountId := uuid.NewString()
		userId := uuid.NewString()
		err := rbacclient.SetupNewAccount(ctx, accountId, testutil.GetTestLogger(t))
		require.NoError(t, err)

		// Set user as job viewer
		err = rbacclient.SetAccountRole(ctx, rbac.NewUserIdEntity(userId), rbac.NewAccountIdEntity(accountId), mgmtv1alpha1.AccountRole_ACCOUNT_ROLE_JOB_VIEWER)
		require.NoError(t, err)

		// Test Account permissions (should only have view)
		err = rbacclient.EnforceAccount(ctx, rbac.NewUserIdEntity(userId), rbac.NewAccountIdEntity(accountId), rbac.AccountAction_View)
		require.NoError(t, err)

		// Test Job permissions (should only have view)
		jobId := uuid.NewString()
		err = rbacclient.EnforceJob(ctx, rbac.NewUserIdEntity(userId), rbac.NewAccountIdEntity(accountId), rbac.NewJobIdEntity(jobId), rbac.JobAction_View)
		require.NoError(t, err)

		err = rbacclient.EnforceJob(ctx, rbac.NewUserIdEntity(userId), rbac.NewAccountIdEntity(accountId), rbac.NewJobIdEntity(jobId), rbac.JobAction_Execute)
		require.Error(t, err)
		require.ErrorContains(t, err, "user does not have permission to execute job")
		err = rbacclient.EnforceJob(ctx, rbac.NewUserIdEntity(userId), rbac.NewAccountIdEntity(accountId), rbac.NewJobIdEntity(jobId), rbac.JobAction_Create)
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
		err = rbacclient.SetAccountRole(ctx, rbac.NewUserIdEntity(userId), rbac.NewAccountIdEntity(accountId), mgmtv1alpha1.AccountRole_ACCOUNT_ROLE_ADMIN)
		require.NoError(t, err)

		err = rbacclient.RemoveAccountRole(ctx, rbac.NewUserIdEntity(userId), rbac.NewAccountIdEntity(accountId), mgmtv1alpha1.AccountRole_ACCOUNT_ROLE_ADMIN)
		require.NoError(t, err)

		// Verify permissions are removed
		err = rbacclient.EnforceAccount(ctx, rbac.NewUserIdEntity(userId), rbac.NewAccountIdEntity(accountId), rbac.AccountAction_View)
		require.Error(t, err)
		require.ErrorContains(t, err, "user does not have permission to view account")
	})

	t.Run("cross_account_access", func(t *testing.T) {
		t.Parallel()

		// Setup first account and user
		account1Id := uuid.NewString()
		user1Id := uuid.NewString()
		err := rbacclient.SetupNewAccount(ctx, account1Id, testutil.GetTestLogger(t))
		require.NoError(t, err)
		err = rbacclient.SetAccountRole(ctx, rbac.NewUserIdEntity(user1Id), rbac.NewAccountIdEntity(account1Id), mgmtv1alpha1.AccountRole_ACCOUNT_ROLE_ADMIN)
		require.NoError(t, err)

		// Setup second account and user
		account2Id := uuid.NewString()
		user2Id := uuid.NewString()
		err = rbacclient.SetupNewAccount(ctx, account2Id, testutil.GetTestLogger(t))
		require.NoError(t, err)
		err = rbacclient.SetAccountRole(ctx, rbac.NewUserIdEntity(user2Id), rbac.NewAccountIdEntity(account2Id), mgmtv1alpha1.AccountRole_ACCOUNT_ROLE_ADMIN)
		require.NoError(t, err)

		// Verify user1 cannot access account2's resources
		err = rbacclient.EnforceAccount(ctx, rbac.NewUserIdEntity(user1Id), rbac.NewAccountIdEntity(account2Id), rbac.AccountAction_View)
		require.Error(t, err)
		require.ErrorContains(t, err, "user does not have permission to view account")

		// Verify user1 cannot access jobs in account2
		jobId := uuid.NewString()
		err = rbacclient.EnforceJob(ctx, rbac.NewUserIdEntity(user1Id), rbac.NewAccountIdEntity(account2Id), rbac.NewJobIdEntity(jobId), rbac.JobAction_View)
		require.Error(t, err)
		require.ErrorContains(t, err, "user does not have permission to view job")

		// Verify user1 cannot access connections in account2
		connectionId := uuid.NewString()
		err = rbacclient.EnforceConnection(ctx, rbac.NewUserIdEntity(user1Id), rbac.NewAccountIdEntity(account2Id), rbac.NewConnectionIdEntity(connectionId), rbac.ConnectionAction_View)
		require.Error(t, err)
		require.ErrorContains(t, err, "user does not have permission to view connection")
	})

	t.Run("mixed_roles_same_account", func(t *testing.T) {
		t.Parallel()

		accountId := uuid.NewString()
		adminUserId := uuid.NewString()
		viewerUserId := uuid.NewString()
		err := rbacclient.SetupNewAccount(ctx, accountId, testutil.GetTestLogger(t))
		require.NoError(t, err)

		// Set up admin user
		err = rbacclient.SetAccountRole(ctx, rbac.NewUserIdEntity(adminUserId), rbac.NewAccountIdEntity(accountId), mgmtv1alpha1.AccountRole_ACCOUNT_ROLE_ADMIN)
		require.NoError(t, err)

		// Set up job viewer user
		err = rbacclient.SetAccountRole(ctx, rbac.NewUserIdEntity(viewerUserId), rbac.NewAccountIdEntity(accountId), mgmtv1alpha1.AccountRole_ACCOUNT_ROLE_JOB_VIEWER)
		require.NoError(t, err)

		jobId := uuid.NewString()
		connectionId := uuid.NewString()

		// Verify admin has full permissions
		err = rbacclient.EnforceJob(ctx, rbac.NewUserIdEntity(adminUserId), rbac.NewAccountIdEntity(accountId), rbac.NewJobIdEntity(jobId), rbac.JobAction_Create)
		require.NoError(t, err)
		err = rbacclient.EnforceJob(ctx, rbac.NewUserIdEntity(adminUserId), rbac.NewAccountIdEntity(accountId), rbac.NewJobIdEntity(jobId), rbac.JobAction_Delete)
		require.NoError(t, err)
		err = rbacclient.EnforceConnection(ctx, rbac.NewUserIdEntity(adminUserId), rbac.NewAccountIdEntity(accountId), rbac.NewConnectionIdEntity(connectionId), rbac.ConnectionAction_Create)
		require.NoError(t, err)

		// Verify job viewer can only view and execute jobs
		err = rbacclient.EnforceJob(ctx, rbac.NewUserIdEntity(viewerUserId), rbac.NewAccountIdEntity(accountId), rbac.NewJobIdEntity(jobId), rbac.JobAction_View)
		require.NoError(t, err)

		// Verify job viewer cannot perform other actions

		err = rbacclient.EnforceJob(ctx, rbac.NewUserIdEntity(viewerUserId), rbac.NewAccountIdEntity(accountId), rbac.NewJobIdEntity(jobId), rbac.JobAction_Execute)
		require.Error(t, err)
		require.ErrorContains(t, err, "user does not have permission to execute job")

		err = rbacclient.EnforceJob(ctx, rbac.NewUserIdEntity(viewerUserId), rbac.NewAccountIdEntity(accountId), rbac.NewJobIdEntity(jobId), rbac.JobAction_Create)
		require.Error(t, err)
		require.ErrorContains(t, err, "user does not have permission to create job")

		err = rbacclient.EnforceJob(ctx, rbac.NewUserIdEntity(viewerUserId), rbac.NewAccountIdEntity(accountId), rbac.NewJobIdEntity(jobId), rbac.JobAction_Delete)
		require.Error(t, err)
		require.ErrorContains(t, err, "user does not have permission to delete job")

		err = rbacclient.EnforceJob(ctx, rbac.NewUserIdEntity(viewerUserId), rbac.NewAccountIdEntity(accountId), rbac.NewJobIdEntity(jobId), rbac.JobAction_Edit)
		require.Error(t, err)
		require.ErrorContains(t, err, "user does not have permission to edit job")
	})

}
