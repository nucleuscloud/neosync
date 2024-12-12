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
	Identifier
	GetAccountId() string
}
type DomainEntityImpl struct {
	id        string
	accountId string
	isWild    bool
}

type Identifier interface {
	GetId() string
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

type IdentifierImpl struct {
	id string
}

func NewIdentifier(id string) Identifier {
	return &IdentifierImpl{
		id: id,
	}
}

func (i *IdentifierImpl) GetId() string {
	return i.id
}

func (u *UserEntityEnforcer) EnforceJob(ctx context.Context, job DomainEntity, action rbac.JobAction) error {
	if err := u.enforceAccountAccess(ctx, job.GetAccountId()); err != nil {
		return err
	}
	return u.enforcer.EnforceJob(ctx, u.user, rbac.NewAccountIdEntity(job.GetAccountId()), rbac.NewJobIdEntity(job.GetId()), action)
}

func (u *UserEntityEnforcer) EnforceConnection(ctx context.Context, connection DomainEntity, action rbac.ConnectionAction) error {
	if err := u.enforceAccountAccess(ctx, connection.GetAccountId()); err != nil {
		return err
	}
	return u.enforcer.EnforceConnection(ctx, u.user, rbac.NewAccountIdEntity(connection.GetAccountId()), rbac.NewConnectionIdEntity(connection.GetId()), action)
}

func (u *UserEntityEnforcer) EnforceAccount(ctx context.Context, account Identifier, action rbac.AccountAction) error {
	if err := u.enforceAccountAccess(ctx, account.GetId()); err != nil {
		return err
	}
	return u.enforcer.EnforceAccount(ctx, u.user, rbac.NewAccountIdEntity(account.GetId()), action)
}
