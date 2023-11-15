package tokenctx

import (
	"context"
	"testing"

	auth_apikey "github.com/nucleuscloud/neosync/backend/internal/auth/apikey"
	auth_jwt "github.com/nucleuscloud/neosync/backend/internal/auth/jwt"
	"github.com/zeebo/assert"
)

func Test_GetTokenCtx_ApiKey(t *testing.T) {
	ctx := context.WithValue(
		context.Background(),
		auth_apikey.TokenContextKey{},
		&auth_apikey.TokenContextData{},
	)
	resp, err := GetTokenCtx(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.ApiKeyContextData)
}

func Test_GetTokenCtx_Jwt(t *testing.T) {
	ctx := context.WithValue(
		context.Background(),
		auth_jwt.TokenContextKey{},
		&auth_jwt.TokenContextData{},
	)
	resp, err := GetTokenCtx(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.JwtContextData)
}

func Test_GetTokenCtx_None(t *testing.T) {
	resp, err := GetTokenCtx(context.Background())
	assert.Error(t, err)
	assert.Nil(t, resp)
}
