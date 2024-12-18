package userdata

import (
	"context"

	"github.com/nucleuscloud/neosync/backend/internal/ee/rbac"
)

type UserEntityEnforcer struct {
	enforcer             rbac.EntityEnforcer
	user                 rbac.EntityString
	enforceAccountAccess func(ctx context.Context, accountId string) error
	isApiKey             bool
}

var _ EntityEnforcer = (*UserEntityEnforcer)(nil)

// Higher level entity enforcer that slightly abstracts away the lower level rbac interface
// The intention here is to be able to use objects that are closer to the mgmt domain model
// rather than the lower level rbac model
// see the entity.go file for functions that help with this
type EntityEnforcer interface {
	EnforceJob(ctx context.Context, job DomainEntity, action rbac.JobAction) error
	EnforceConnection(ctx context.Context, connection DomainEntity, action rbac.ConnectionAction) error
	EnforceAccount(ctx context.Context, account Identifier, action rbac.AccountAction) error
}

func (u *UserEntityEnforcer) EnforceJob(ctx context.Context, job DomainEntity, action rbac.JobAction) error {
	if err := u.enforceAccountAccess(ctx, job.GetAccountId()); err != nil {
		return err
	}
	if u.isApiKey {
		return nil
	}
	return u.enforcer.EnforceJob(ctx, u.user, rbac.NewAccountIdEntity(job.GetAccountId()), rbac.NewJobIdEntity(job.GetId()), action)
}

func (u *UserEntityEnforcer) EnforceConnection(ctx context.Context, connection DomainEntity, action rbac.ConnectionAction) error {
	if err := u.enforceAccountAccess(ctx, connection.GetAccountId()); err != nil {
		return err
	}
	if u.isApiKey {
		return nil
	}
	return u.enforcer.EnforceConnection(ctx, u.user, rbac.NewAccountIdEntity(connection.GetAccountId()), rbac.NewConnectionIdEntity(connection.GetId()), action)
}

func (u *UserEntityEnforcer) EnforceAccount(ctx context.Context, account Identifier, action rbac.AccountAction) error {
	if err := u.enforceAccountAccess(ctx, account.GetId()); err != nil {
		return err
	}
	if u.isApiKey {
		return nil
	}
	return u.enforcer.EnforceAccount(ctx, u.user, rbac.NewAccountIdEntity(account.GetId()), action)
}
