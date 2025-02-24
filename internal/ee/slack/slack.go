package ee_slack

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	sym_encrypt "github.com/nucleuscloud/neosync/internal/encrypt/sym"
	"github.com/slack-go/slack"
)

type Interface interface {
	GetAuthorizeUrl(accountId, userId string) (string, error)
	ValidateState(state, accountId, userId string) error
	ExchangeCodeForAccessToken(ctx context.Context, code string) (*slack.OAuthV2Response, error)
	Test(ctx context.Context, accessToken string) (*slack.AuthTestResponse, error)
	SendMessage(ctx context.Context, accessToken, channelId string, options ...slack.MsgOption) error
}

type Client struct {
	cfg       *config
	encryptor sym_encrypt.Interface
}

type config struct {
	appClientId     string
	appClientSecret string
	scope           string
	redirectUrl     string
}

type Option func(*config)

func WithAppClientId(appClientId string) Option {
	return func(c *config) {
		c.appClientId = appClientId
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

func NewClient(encryptor sym_encrypt.Interface, opts ...Option) *Client {
	cfg := &config{}
	for _, opt := range opts {
		opt(cfg)
	}
	return &Client{cfg: cfg, encryptor: encryptor}
}

func (c *Client) GetAuthorizeUrl(accountId, userId string) (string, error) {
	state := &slackOauthState{
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
			"client_id":    {c.cfg.appClientId},
			"redirect_uri": {c.cfg.redirectUrl},
			"state":        {stateEncrypted},
		}.Encode(),
	}

	return slackUrl.String(), nil
}

func (c *Client) ValidateState(state, accountId, userId string) error {
	stateDecrypted, err := c.encryptor.Decrypt(state)
	if err != nil {
		return fmt.Errorf("unable to decrypt slack oauth state: %w", err)
	}

	var decodedState slackOauthState
	if err := json.Unmarshal([]byte(stateDecrypted), &decodedState); err != nil {
		return fmt.Errorf("unable to unmarshal slack oauth state: %w", err)
	}

	if decodedState.AccountId != accountId {
		return fmt.Errorf("invalid account id")
	}

	if decodedState.UserId != userId {
		return fmt.Errorf("invalid user id")
	}

	if time.Now().Unix()-decodedState.Timestamp > 900 {
		return fmt.Errorf("oauth state expired")
	}
	return nil
}

func (c *Client) ExchangeCodeForAccessToken(ctx context.Context, code string) (*slack.OAuthV2Response, error) {
	resp, err := slack.GetOAuthV2ResponseContext(ctx, &http.Client{}, c.cfg.appClientId, c.cfg.appClientSecret, code, c.cfg.redirectUrl)
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

func (c *Client) SendMessage(ctx context.Context, accessToken, channelId string, options ...slack.MsgOption) error {
	api := slack.New(accessToken)
	_, _, err := api.PostMessageContext(ctx, channelId, options...)
	if err != nil {
		return fmt.Errorf("unable to send message: %w", err)
	}
	return nil
}

type slackOauthState struct {
	AccountId string `json:"accountId"`
	UserId    string `json:"userId"`
	Timestamp int64  `json:"timestamp"`
}
