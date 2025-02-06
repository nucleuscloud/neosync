package keycloak

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/nucleuscloud/neosync/internal/authmgmt"
)

var (
	DefaultTokenExpirationBuffer = 10 * time.Second
)

type TokenProvider interface {
	GetToken(context.Context) (string, error)
}

var _ authmgmt.Interface = &AdminClient{}

type AdminClient struct {
	domain        string
	tokenprovider TokenProvider
	logger        *slog.Logger
}

// there is more here but leaving it out as it's not relevant right now
type user struct {
	Id            string `json:"id"`
	Username      string `json:"username"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"emailVerified"`
	Enabled       bool   `json:"enabled"`
	Firstname     string `json:"firstName"`
	Lastname      string `json:"lastName"`
}

func (u *user) Name() string {
	return fmt.Sprintf("%s %s", u.Firstname, u.Lastname)
}

func New(
	domain string,
	provider TokenProvider,
	logger *slog.Logger,
) (*AdminClient, error) {
	return &AdminClient{domain: domain, tokenprovider: provider, logger: logger}, nil
}

func (c *AdminClient) GetUserBySub(ctx context.Context, sub string) (*authmgmt.User, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/users/%s", c.domain, sub), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("unable to initiate request to user endpoint: %w", err)
	}
	token, err := c.tokenprovider.GetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get access token: %w", err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve response when requested access to user data: %w", err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body for user endpoint: %w", err)
	}

	if res.StatusCode > 399 {
		c.logger.Error("unable to retrieve user", "body", string(body), "statuscode", res.StatusCode)
		return nil, fmt.Errorf("received unsuccessful status code when retrieving keycloak user. code: %d", res.StatusCode)
	}

	var kcuser *user
	err = json.Unmarshal(body, &kcuser)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal keycloak user data: %w", err)
	}
	return &authmgmt.User{
		Name:    kcuser.Name(),
		Email:   kcuser.Email,
		Picture: "",
	}, nil
}
