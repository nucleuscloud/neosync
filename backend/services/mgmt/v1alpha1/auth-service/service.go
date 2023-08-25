package v1alpha1_authservice

import "github.com/nucleuscloud/neosync/backend/internal/auth"

type AuthClient interface {
	GetRefreshedAccessToken(clientId string, refreshToken string) (*auth.AuthTokenResponse, error)
	GetTokenResponse(
		clientId string,
		code string,
		redirectUri string,
	) (*auth.AuthTokenResponse, error)
}

type Service struct {
	cfg *Config

	authclient AuthClient
}

type Config struct{}

func New(
	cfg *Config,
	authclient AuthClient,
) *Service {

	return &Service{
		cfg:        cfg,
		authclient: authclient,
	}
}
