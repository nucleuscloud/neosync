package rbac

import (
	"context"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
)

type AllowAllClient struct {
}

var _ Interface = (*AllowAllClient)(nil)

func (a *AllowAllClient) Job(ctx context.Context, user, account, job EntityString, action JobAction) (bool, error) {
	return true, nil
}

func (a *AllowAllClient) Connection(ctx context.Context, user, account, connection EntityString, action ConnectionAction) (bool, error) {
	return true, nil
}

func (a *AllowAllClient) Account(ctx context.Context, user, account EntityString, action AccountAction) (bool, error) {
	return true, nil
}

func (a *AllowAllClient) EnforceJob(ctx context.Context, user, account, job EntityString, action JobAction) error {
	return nil
}

func (a *AllowAllClient) EnforceConnection(ctx context.Context, user, account, connection EntityString, action ConnectionAction) error {
	return nil
}

func (a *AllowAllClient) EnforceAccount(ctx context.Context, user, account EntityString, action AccountAction) error {
	return nil
}

func (a *AllowAllClient) SetAccountRole(ctx context.Context, user, account EntityString, role mgmtv1alpha1.AccountRole) error {
	return nil
}

func (a *AllowAllClient) RemoveAccountRole(ctx context.Context, user, account EntityString, role mgmtv1alpha1.AccountRole) error {
	return nil
}

func (a *AllowAllClient) RemoveAccountUser(ctx context.Context, user, account EntityString) error {
	return nil
}

func NewAllowAllClient() *AllowAllClient {
	return &AllowAllClient{}
}
