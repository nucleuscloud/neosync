package auth0

import (
	"context"

	"github.com/auth0/go-auth0/management"
	"github.com/nucleuscloud/neosync/internal/authmgmt"
)

var _ authmgmt.Interface = &Auth0MgmtClient{} // ensures it always conforms to the interface

type Auth0MgmtClient struct {
	client *management.Management
}

func New(domain, clientId, clientSecret string) (*Auth0MgmtClient, error) {
	client, err := management.New(domain, management.WithClientCredentials(context.Background(), clientId, clientSecret))
	if err != nil {
		return nil, err
	}
	return &Auth0MgmtClient{
		client: client,
	}, nil
}

func (c *Auth0MgmtClient) GetUserBySub(ctx context.Context, id string) (*authmgmt.User, error) {
	user, err := c.client.User.Read(ctx, id)
	if err != nil {
		return nil, err
	}
	return &authmgmt.User{
		Name:    user.GetName(),
		Email:   user.GetEmail(),
		Picture: user.GetPicture(),
	}, nil
}
