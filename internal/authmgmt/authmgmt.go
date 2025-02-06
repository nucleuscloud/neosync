package authmgmt

import (
	"context"
	"errors"
)

type User struct {
	Name    string
	Email   string
	Picture string
}

type Interface interface {
	GetUserBySub(ctx context.Context, sub string) (*User, error)
}

type UnimplementedClient struct{}

var _ Interface = &UnimplementedClient{}

func (c *UnimplementedClient) GetUserBySub(ctx context.Context, sub string) (*User, error) {
	return nil, errors.ErrUnsupported
}
