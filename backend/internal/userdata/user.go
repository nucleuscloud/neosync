package userdata

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgtype"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/internal/apikey"
	auth_apikey "github.com/nucleuscloud/neosync/backend/internal/auth/apikey"
	"github.com/nucleuscloud/neosync/backend/internal/ee/rbac"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
)

type UserAccountServiceClient interface {
	IsUserInAccount(ctx context.Context, req *connect.Request[mgmtv1alpha1.IsUserInAccountRequest]) (*connect.Response[mgmtv1alpha1.IsUserInAccountResponse], error)
}

type User struct {
	id pgtype.UUID

	apiKeyData *auth_apikey.TokenContextData // Optional because we may not have an api key

	userAccountServiceClient UserAccountServiceClient

	UserEntityEnforcer
}

func (u *User) Id() string {
	return neosyncdb.UUIDString(u.id)
}
func (u *User) PgId() pgtype.UUID {
	return u.id
}

func (u *User) IsWorkerApiKey() bool {
	return u.apiKeyData != nil && u.apiKeyData.ApiKeyType == apikey.WorkerApiKey
}

func (u *User) IsApiKey() bool {
	return u.apiKeyData != nil
}

// type ValidateAccountAccessResponse struct {
// 	accountId pgtype.UUID
// }

// func (v *ValidateAccountAccessResponse) AccountId() string {
// 	return neosyncdb.UUIDString(v.accountId)
// }
// func (v *ValidateAccountAccessResponse) PgAccountId() pgtype.UUID {
// 	return v.accountId
// }

func EnforceAccountAccess(ctx context.Context, user *User, accountId string) error {
	if user.IsApiKey() {
		if user.IsWorkerApiKey() {
			return nil
		}
		// We first want to check to make sure the api key is valid and that it says it's in the account
		// However, we still want to make a DB request to ensure the DB still says it's in the account
		if user.apiKeyData.ApiKey == nil || neosyncdb.UUIDString(user.apiKeyData.ApiKey.AccountID) != accountId {
			return nucleuserrors.NewForbidden("api key is not valid for account")
		}
	}

	// Note: because we are calling to the user account service here, the ctx must still contain the user data
	inAccountResp, err := user.userAccountServiceClient.IsUserInAccount(ctx, connect.NewRequest(&mgmtv1alpha1.IsUserInAccountRequest{AccountId: accountId}))
	if err != nil {
		return fmt.Errorf("unable to check if user is in account: %w", err)
	}
	if !inAccountResp.Msg.GetOk() {
		return nucleuserrors.NewForbidden("user is not in account")
	}
	return nil
}

//	type UserEntityEnforcer interface {
//		Job(ctx context.Context, account rbac.EntityString, job rbac.EntityString, action rbac.JobAction) (bool, error)
//		Connection(ctx context.Context, account rbac.EntityString, connection rbac.EntityString, action rbac.ConnectionAction) (bool, error)
//		Account(ctx context.Context, account rbac.EntityString, action rbac.AccountAction) (bool, error)
//	}
type UserEntityEnforcer struct {
	enforcer             rbac.EntityEnforcer
	user                 rbac.EntityString
	enforceAccountAccess func(ctx context.Context, accountId string) error
}

type DomainEntity interface {
	GetId() string
	GetAccountId() string
}
type DomainEntityImpl struct {
	id        string
	accountId string
	isWild    bool
}

func (j *DomainEntityImpl) GetId() string {
	return j.id
}
func (j *DomainEntityImpl) GetAccountId() string {
	return j.accountId
}

func NewDomainEntity(accountId, id string) DomainEntity {
	return &DomainEntityImpl{
		id:        id,
		accountId: accountId,
	}
}

func NewWildcardDomainEntity(accountId string) DomainEntity {
	return &DomainEntityImpl{
		id:        rbac.Wildcard,
		accountId: accountId,
		isWild:    true,
	}
}

func NewDbDomainEntity(accountId, id pgtype.UUID) DomainEntity {
	return &DomainEntityImpl{
		id:        neosyncdb.UUIDString(id),
		accountId: neosyncdb.UUIDString(accountId),
	}
}

func (u *UserEntityEnforcer) Job(ctx context.Context, job DomainEntity, action rbac.JobAction) (bool, error) {
	if err := u.enforceAccountAccess(ctx, job.GetAccountId()); err != nil {
		return false, err
	}
	return u.enforcer.Job(ctx, u.user, rbac.NewAccountIdEntity(job.GetAccountId()), rbac.NewJobIdEntity(job.GetId()), action)
}
func (u *UserEntityEnforcer) EnforceJob(ctx context.Context, job DomainEntity, action rbac.JobAction) error {
	ok, err := u.Job(ctx, job, action)
	if err != nil {
		return err
	}
	if !ok {
		return nucleuserrors.NewForbidden(fmt.Sprintf("user does not have permission to %s job", action))
	}
	return nil
}
func (u *UserEntityEnforcer) Connection(ctx context.Context, connection DomainEntity, action rbac.ConnectionAction) (bool, error) {
	if err := u.enforceAccountAccess(ctx, connection.GetAccountId()); err != nil {
		return false, err
	}
	return u.enforcer.Connection(ctx, u.user, rbac.NewAccountIdEntity(connection.GetAccountId()), rbac.NewConnectionIdEntity(connection.GetId()), action)
}
func (u *UserEntityEnforcer) EnforceConnection(ctx context.Context, connection DomainEntity, action rbac.ConnectionAction) error {
	ok, err := u.Connection(ctx, connection, action)
	if err != nil {
		return err
	}
	if !ok {
		return nucleuserrors.NewForbidden(fmt.Sprintf("user does not have permission to %s connection", action))
	}
	return nil
}
func (u *UserEntityEnforcer) Account(ctx context.Context, account *mgmtv1alpha1.UserAccount, action rbac.AccountAction) (bool, error) {
	if err := u.enforceAccountAccess(ctx, account.GetId()); err != nil {
		return false, err
	}
	return u.enforcer.Account(ctx, u.user, rbac.NewAccountIdEntity(account.GetId()), action)
}
func (u *UserEntityEnforcer) EnforceAccount(ctx context.Context, account *mgmtv1alpha1.UserAccount, action rbac.AccountAction) error {
	ok, err := u.Account(ctx, account, action)
	if err != nil {
		return err
	}
	if !ok {
		return nucleuserrors.NewForbidden(fmt.Sprintf("user does not have permission to %s account", action))
	}
	return nil
}
