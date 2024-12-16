package userdata

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	auth_apikey "github.com/nucleuscloud/neosync/backend/internal/auth/apikey"
	"github.com/nucleuscloud/neosync/backend/internal/ee/rbac"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
)

type UserServiceClient interface {
	GetUser(ctx context.Context, req *connect.Request[mgmtv1alpha1.GetUserRequest]) (*connect.Response[mgmtv1alpha1.GetUserResponse], error)
	IsUserInAccount(ctx context.Context, req *connect.Request[mgmtv1alpha1.IsUserInAccountRequest]) (*connect.Response[mgmtv1alpha1.IsUserInAccountResponse], error)
}

type Client struct {
	userServiceClient UserServiceClient
	enforcer          rbac.EntityEnforcer
}

type Interface interface {
	GetUser(ctx context.Context) (*User, error)
}

type GetUserResponse struct {
	User *User
}

func NewClient(
	userServiceClient UserServiceClient,
	enforcer rbac.EntityEnforcer,
) *Client {
	return &Client{
		userServiceClient: userServiceClient,
		enforcer:          enforcer,
	}
}

func (c *Client) GetUser(ctx context.Context) (*User, error) {
	resp, err := c.userServiceClient.GetUser(ctx, connect.NewRequest(&mgmtv1alpha1.GetUserRequest{}))
	if err != nil {
		return nil, fmt.Errorf("unable to get user: %w", err)
	}
	pguuid, err := neosyncdb.ToUuid(resp.Msg.GetUserId())
	if err != nil {
		return nil, fmt.Errorf("unable to parse user id: %w", err)
	}

	apiKeyData, _ := auth_apikey.GetTokenDataFromCtx(ctx)

	user := &User{
		id:                       pguuid,
		apiKeyData:               apiKeyData,
		userAccountServiceClient: c.userServiceClient,
	}
	user.EntityEnforcer = &UserEntityEnforcer{
		enforcer: c.enforcer,
		user:     rbac.NewUserIdEntity(resp.Msg.GetUserId()),
		enforceAccountAccess: func(ctx context.Context, accountId string) error {
			return EnforceAccountAccess(ctx, user, accountId)
		},
	}

	return user, nil
}
