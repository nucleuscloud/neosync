package auth0

import (
	"context"

	"github.com/auth0/go-auth0/management"
)

type Auth0MgmtClientInterface interface {
	GetUserById(ctx context.Context, id string) (*management.User, error)
}

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

func (c *Auth0MgmtClient) GetUserById(ctx context.Context, id string) (*management.User, error) {
	return c.client.User.Read(ctx, id)
}
