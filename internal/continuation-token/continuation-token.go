package continuation_token

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

func FromTokenString(tokenStr string) (*ContinuationToken, error) {
	if tokenStr == "" {
		return nil, nil
	}

	bytes, err := base64.StdEncoding.DecodeString(tokenStr)
	if err != nil {
		return nil, fmt.Errorf("unable to decode continuation token: %w", err)
	}

	var token ContinuationToken
	if err := json.Unmarshal(bytes, &token); err != nil {
		return nil, fmt.Errorf("unable to unmarshal continuation token: %w", err)
	}

	return &token, nil
}

type ContinuationToken struct {
	Contents *Contents `json:"contents"`
}
type Contents struct {
	LastReadOrderValues []any `json:"lastReadOrderValues"`
}

func NewContents(lastReadOrderValues []any) *Contents {
	return &Contents{
		LastReadOrderValues: lastReadOrderValues,
	}
}

func NewFromContents(contents *Contents) *ContinuationToken {
	return &ContinuationToken{
		Contents: contents,
	}
}

func (c *ContinuationToken) String() string {
	bytes, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(bytes)
}
