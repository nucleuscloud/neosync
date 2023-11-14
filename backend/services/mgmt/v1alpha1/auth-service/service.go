package v1alpha1_authservice

import (
	"context"

	auth_client "github.com/nucleuscloud/neosync/backend/internal/auth/client"
	"github.com/nucleuscloud/neosync/backend/internal/authmgmt/auth0"
)

type Service struct {
	cfg        *Config
	authclient AuthClient
	auth0Mgmt  *auth0.Auth0MgmtClient
}

type Config struct {
	IsAuthEnabled bool

	CliClientId  string
	CliAudience  string
	IssuerUrl    string
	AuthorizeUrl string
}

type AuthClient interface {
	GetTokenResponse(
		ctx context.Context,
		clientId string,
		code string,
		redirecturi string,
	) (*auth_client.AuthTokenResponse, error)
}

func New(
	cfg *Config,
	authclient AuthClient,
	auth0Mgmt *auth0.Auth0MgmtClient,
) *Service {
	return &Service{cfg: cfg, authclient: authclient, auth0Mgmt: auth0Mgmt}
}
