package authmgmt

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_UnimplementedClient_GetUserBySub(t *testing.T) {
	client := &UnimplementedClient{}
	user, err := client.GetUserBySub(context.Background(), "foo")
	assert.Nil(t, user)
	assert.ErrorIs(t, err, errors.ErrUnsupported)
}
