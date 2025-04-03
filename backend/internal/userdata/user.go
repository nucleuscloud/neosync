package userdata

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgtype"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	auth_apikey "github.com/nucleuscloud/neosync/backend/internal/auth/apikey"
	"github.com/nucleuscloud/neosync/internal/apikey"
	"github.com/nucleuscloud/neosync/internal/ee/license"
	nucleuserrors "github.com/nucleuscloud/neosync/internal/errors"
	"github.com/nucleuscloud/neosync/internal/neosyncdb"
)

type UserAccountServiceClient interface {
	IsUserInAccount(
		ctx context.Context,
		req *connect.Request[mgmtv1alpha1.IsUserInAccountRequest],
	) (*connect.Response[mgmtv1alpha1.IsUserInAccountResponse], error)
}

type User struct {
	id pgtype.UUID

	apiKeyData *auth_apikey.TokenContextData // Optional because we may not have an api key

	userAccountServiceClient UserAccountServiceClient

	EntityEnforcer

	license license.EEInterface
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

func (u *User) EnforceAccountAccess(ctx context.Context, accountId string) error {
	return enforceAccountAccess(ctx, u, accountId)
}

func (u *User) EnforceLicense(ctx context.Context, accountId string) error {
	ok, err := u.IsLicensed(ctx, accountId)
	if err != nil {
		return err
	}
	if !ok {
		return nucleuserrors.NewUnauthorized("account does not have an active license")
	}
	return nil
}

func (u *User) IsLicensed(ctx context.Context, accountId string) (bool, error) {
	if err := u.EnforceAccountAccess(ctx, accountId); err != nil {
		return false, err
	}

	// todo: check account type for Neosync Cloud Cloud?
	// if: personal, then check if free trial is active
	// if: pro, then no? or maybe still do a trial check?
	// if: enterprise, then check for valid license

	if u.license == nil {
		return false, nil
	}

	return u.license.IsValid(), nil
}

func enforceAccountAccess(ctx context.Context, user *User, accountId string) error {
	if user.IsApiKey() {
		if user.IsWorkerApiKey() {
			return nil
		}
		// We first want to check to make sure the api key is valid and that it says it's in the account
		// However, we still want to make a DB request to ensure the DB still says it's in the account
		if user.apiKeyData.ApiKey == nil ||
			neosyncdb.UUIDString(user.apiKeyData.ApiKey.AccountID) != accountId {
			return nucleuserrors.NewUnauthorized("api key is not valid for account")
		}
	}

	// Note: because we are calling to the user account service here, the ctx must still contain the user data
	inAccountResp, err := user.userAccountServiceClient.IsUserInAccount(
		ctx,
		connect.NewRequest(&mgmtv1alpha1.IsUserInAccountRequest{AccountId: accountId}),
	)
	if err != nil {
		return fmt.Errorf("unable to check if user is in account: %w", err)
	}
	if !inAccountResp.Msg.GetOk() {
		return nucleuserrors.NewUnauthorized("user is not in account")
	}
	return nil
}
