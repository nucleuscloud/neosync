package ee_slack

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	sym_encrypt "github.com/nucleuscloud/neosync/internal/encrypt/sym"
	"github.com/slack-go/slack"
)

type IsUserInAccountFunc func(ctx context.Context, userId, accountId string) (bool, error)

type Interface interface {
	GetAuthorizeUrl(accountId, userId string) (string, error)
	ValidateState(
		ctx context.Context,
		state, userId string,
		isUserInAccount IsUserInAccountFunc,
	) (*OauthState, error)
	ExchangeCodeForAccessToken(ctx context.Context, code string) (*slack.OAuthV2Response, error)
	Test(ctx context.Context, accessToken string) (*slack.AuthTestResponse, error)
	SendMessage(
		ctx context.Context,
		accessToken, channelId string,
		options ...slack.MsgOption,
	) error
	JoinChannel(ctx context.Context, accessToken, channelId string, logger *slog.Logger) error
	GetPublicChannels(ctx context.Context, accessToken string) ([]slack.Channel, error)
}

type Client struct {
	cfg       *config
	encryptor sym_encrypt.Interface
}

type config struct {
	authClientId     string
	authClientSecret string
	scope            string
	redirectUrl      string

	httpClient *http.Client
}

type Option func(*config)

func WithAuthClientCreds(authClientId, authClientSecret string) Option {
	return func(c *config) {
		c.authClientId = authClientId
		c.authClientSecret = authClientSecret
	}
}

func WithScope(scope string) Option {
	return func(c *config) {
		c.scope = scope
	}
}

func WithRedirectUrl(redirectUrl string) Option {
	return func(c *config) {
		c.redirectUrl = redirectUrl
	}
}

func WithHttpClient(httpClient *http.Client) Option {
	return func(c *config) {
		c.httpClient = httpClient
	}
}

func NewClient(encryptor sym_encrypt.Interface, opts ...Option) *Client {
	cfg := &config{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return &Client{cfg: cfg, encryptor: encryptor}
}

func (c *Client) GetAuthorizeUrl(accountId, userId string) (string, error) {
	state := &OauthState{
		AccountId: accountId,
		UserId:    userId,
		Timestamp: time.Now().Unix(),
	}

	stateBits, err := json.Marshal(state)
	if err != nil {
		return "", fmt.Errorf("unable to marshal slack oauth state: %w", err)
	}

	stateEncrypted, err := c.encryptor.Encrypt(string(stateBits))
	if err != nil {
		return "", fmt.Errorf("unable to encrypt slack oauth state: %w", err)
	}

	slackUrl := &url.URL{
		Scheme: "https",
		Host:   "slack.com",
		Path:   "/oauth/v2/authorize",
		RawQuery: url.Values{
			"scope":        {c.cfg.scope},
			"client_id":    {c.cfg.authClientId},
			"redirect_uri": {c.cfg.redirectUrl},
			"state":        {stateEncrypted},
		}.Encode(),
	}

	return slackUrl.String(), nil
}

func (c *Client) ValidateState(
	ctx context.Context,
	state, userId string,
	isUserInAccount IsUserInAccountFunc,
) (*OauthState, error) {
	stateDecrypted, err := c.encryptor.Decrypt(state)
	if err != nil {
		return nil, fmt.Errorf("unable to decrypt slack oauth state: %w", err)
	}

	var decodedState OauthState
	if err := json.Unmarshal([]byte(stateDecrypted), &decodedState); err != nil {
		return nil, fmt.Errorf("unable to unmarshal slack oauth state: %w", err)
	}

	if decodedState.UserId != userId {
		return nil, fmt.Errorf("invalid user id")
	}

	userInAccount, err := isUserInAccount(ctx, userId, decodedState.AccountId)
	if err != nil {
		return nil, fmt.Errorf("unable to check if user is in account: %w", err)
	}

	if !userInAccount {
		return nil, fmt.Errorf("invalid account id")
	}

	if time.Now().Unix()-decodedState.Timestamp > 900 {
		return nil, fmt.Errorf("oauth state expired")
	}
	return &decodedState, nil
}

func (c *Client) ExchangeCodeForAccessToken(
	ctx context.Context,
	code string,
) (*slack.OAuthV2Response, error) {
	resp, err := slack.GetOAuthV2ResponseContext(
		ctx,
		c.cfg.httpClient,
		c.cfg.authClientId,
		c.cfg.authClientSecret,
		code,
		c.cfg.redirectUrl,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to exchange code for access token: %w", err)
	}
	if resp.Err() != nil {
		return nil, fmt.Errorf("unable to exchange code for access token: %w", resp.Err())
	}
	return resp, nil
}

func (c *Client) Test(ctx context.Context, accessToken string) (*slack.AuthTestResponse, error) {
	api := slack.New(accessToken)

	resp, err := api.AuthTestContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to test slack connection: %w", err)
	}
	return resp, nil
}

func (c *Client) SendMessage(
	ctx context.Context,
	accessToken, channelId string,
	options ...slack.MsgOption,
) error {
	api := slack.New(accessToken)
	_, _, err := api.PostMessageContext(ctx, channelId, options...)
	if err != nil {
		return fmt.Errorf("unable to send message: %w", err)
	}
	return nil
}

func (c *Client) JoinChannel(
	ctx context.Context,
	accessToken, channelId string,
	logger *slog.Logger,
) error {
	api := slack.New(accessToken)

	_, _, warnings, err := api.JoinConversationContext(ctx, channelId)
	if err != nil {
		return fmt.Errorf("unable to join channel: %w", err)
	}
	if len(warnings) > 0 {
		logger.Warn("warnings when joining slack channel", "warnings", warnings)
	}
	return nil
}

func (c *Client) GetPublicChannels(
	ctx context.Context,
	accessToken string,
) ([]slack.Channel, error) {
	api := slack.New(accessToken)

	channels, _, err := api.GetConversationsContext(ctx, &slack.GetConversationsParameters{
		Limit: 200,
		Types: []string{"public_channel"},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get channels: %w", err)
	}
	return channels, nil
}

type OauthState struct {
	AccountId string `json:"accountId"`
	UserId    string `json:"userId"`
	Timestamp int64  `json:"timestamp"`
}
