package authmw

import (
	"context"
	"errors"
	"net/http"
	"testing"

	auth_apikey "github.com/nucleuscloud/neosync/backend/internal/auth/apikey"
	"github.com/stretchr/testify/mock"
	"github.com/zeebo/assert"
)

func Test_New(t *testing.T) {
	mockAuthClient := NewMockAuthClient(t)
	mw := New(mockAuthClient, mockAuthClient)
	assert.NotNil(t, mw)
}

func Test_AuthMiddleware_InjectTokenCtx_ApiKey(t *testing.T) {
	mockJwt := NewMockAuthClient(t)
	mockApiKey := NewMockAuthClient(t)

	mw := New(mockJwt, mockApiKey)

	mockApiKey.On("InjectTokenCtx", mock.Anything, mock.Anything).
		Return(context.Background(), nil)

	_, err := mw.InjectTokenCtx(context.Background(), http.Header{
		"Authorization": []string{"Bearer foo"},
	})
	assert.NoError(t, err)
}

func Test_AuthMiddleware_InjectTokenCtx_ApiKey_InternalError(t *testing.T) {
	mockJwt := NewMockAuthClient(t)
	mockApiKey := NewMockAuthClient(t)

	mw := New(mockJwt, mockApiKey)

	mockApiKey.On("InjectTokenCtx", mock.Anything, mock.Anything).
		Return(nil, errors.New("internal"))

	_, err := mw.InjectTokenCtx(context.Background(), http.Header{
		"Authorization": []string{"Bearer foo"},
	})
	assert.Error(t, err)
}

func Test_AuthMiddleware_InjectTokenCtx_ApiKey_JwtFallback(t *testing.T) {
	mockJwt := NewMockAuthClient(t)
	mockApiKey := NewMockAuthClient(t)

	mw := New(mockJwt, mockApiKey)

	mockApiKey.On("InjectTokenCtx", mock.Anything, mock.Anything).
		Return(nil, auth_apikey.InvalidApiKeyErr)
	mockJwt.On("InjectTokenCtx", mock.Anything, mock.Anything).
		Return(context.Background(), nil)

	_, err := mw.InjectTokenCtx(context.Background(), http.Header{
		"Authorization": []string{"Bearer foo"},
	})
	assert.NoError(t, err)
}
