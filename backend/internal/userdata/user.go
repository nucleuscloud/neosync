package userdata

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgtype"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/internal/apikey"
	auth_apikey "github.com/nucleuscloud/neosync/backend/internal/auth/apikey"
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

	EntityEnforcer
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
