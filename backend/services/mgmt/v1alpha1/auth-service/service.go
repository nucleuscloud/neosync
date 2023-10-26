package v1alpha1_authservice

import (
	"context"

	auth_client "github.com/nucleuscloud/neosync/backend/internal/auth/client"
)

type Service struct {
	cfg        *Config
	authclient AuthClient
}

type Config struct {
	IsAuthEnabled bool

	CliClientId  string
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
) *Service {
	return &Service{cfg: cfg, authclient: authclient}
}
