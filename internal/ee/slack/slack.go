package ee_slack

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	sym_encrypt "github.com/nucleuscloud/neosync/internal/encrypt/sym"
)

type Interface interface {
	GetAuthorizeUrl(accountId, userId string) (string, error)
	ValidateState(state, accountId, userId string) error
	ExchangeCodeForAccessToken(code string) (string, error)
}

type Client struct {
	cfg       *config
	encryptor sym_encrypt.Interface
}

type config struct {
	appClientId string
	scope       string
	redirectUrl string
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

func (c *Client) ExchangeCodeForAccessToken(code string) (string, error) {
	return "todo", nil
}

type slackOauthState struct {
	AccountId string `json:"accountId"`
	UserId    string `json:"userId"`
	Timestamp int64  `json:"timestamp"`
}
