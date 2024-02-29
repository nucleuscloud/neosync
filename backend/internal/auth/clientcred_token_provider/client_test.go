package clientcredtokenprovider

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	auth_client "github.com/nucleuscloud/neosync/backend/internal/auth/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/sync/errgroup"
)

func Test_New(t *testing.T) {
	provider := New("", "", "", 1, slog.Default())
	assert.NotNil(t, provider)
}

func Test_ClientCredentialsTokenProvider_GetToken(t *testing.T) {
	mockTokenProvider := NewMocktokenProvider(t)
	mockTokenProvider.On("GetToken", mock.Anything).Return(&auth_client.AuthTokenResponse{
		Result: &auth_client.AuthTokenResponseData{
			AccessToken: "test-token",
		},
	}, nil)
	provider := &ClientCredentialsTokenProvider{
		logger:         slog.Default(),
		tokenExpBuffer: 0,
		tokenprovider:  mockTokenProvider,
	}
	token, err := provider.GetToken(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, "test-token", token)
}

func Test_ClientCredentialsTokenProvider_GetToken_Cached(t *testing.T) {
	accessToken := "test-token"
	expiresAt := time.Now().Add(1 * time.Minute)
	provider := &ClientCredentialsTokenProvider{
		logger:         slog.Default(),
		tokenExpBuffer: 0,
		tokenprovider:  nil,
		accessToken:    &accessToken,
		expiresAt:      &expiresAt,
	}
	token, err := provider.GetToken(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, "test-token", token)
}

func Test_ClientCredentialsTokenProvider_GetToken_ConcurrentAccess(t *testing.T) {
	mockTokenProvider := NewMocktokenProvider(t)
	mockTokenProvider.On("GetToken", mock.Anything).Return(&auth_client.AuthTokenResponse{
		Result: &auth_client.AuthTokenResponseData{
			AccessToken: "test-token",
			ExpiresIn:   5,
		},
	}, nil)
	provider := &ClientCredentialsTokenProvider{
		logger:         slog.Default(),
		tokenExpBuffer: 0,
		tokenprovider:  mockTokenProvider,
	}

	errgrp, errctx := errgroup.WithContext(context.Background())
	errgrp.Go(func() error {
		_, err := provider.GetToken(errctx)
		return err
	})
	errgrp.Go(func() error {
		_, err := provider.GetToken(errctx)
		return err
	})
	errgrp.Go(func() error {
		_, err := provider.GetToken(errctx)
		return err
	})
	err := errgrp.Wait()
	assert.NoError(t, err)

	token, err := provider.GetToken(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, "test-token", token)
	assert.True(t, mockTokenProvider.AssertNumberOfCalls(t, "GetToken", 1))
}

func Test_ClientCredentialsTokenProvider_GetToken_FailToRetrieveToken(t *testing.T) {
	mockTokenProvider := NewMocktokenProvider(t)
	mockTokenProvider.On("GetToken", mock.Anything).Return(nil, errors.New("test-error"))
	provider := &ClientCredentialsTokenProvider{
		logger:         slog.Default(),
		tokenExpBuffer: 0,
		tokenprovider:  mockTokenProvider,
	}
	token, err := provider.GetToken(context.Background())
	assert.ErrorContains(t, err, "test-error")
	assert.Empty(t, token)
}

func Test_ClientCredentialsTokenProvider_GetToken_FailWithErrorResponse(t *testing.T) {
	mockTokenProvider := NewMocktokenProvider(t)
	mockTokenProvider.On("GetToken", mock.Anything).Return(&auth_client.AuthTokenResponse{
		Error: &auth_client.AuthTokenErrorData{
			Error:            "test-error",
			ErrorDescription: "test-description",
		},
	}, nil)
	provider := &ClientCredentialsTokenProvider{
		logger:         slog.Default(),
		tokenExpBuffer: 0,
		tokenprovider:  mockTokenProvider,
	}
	token, err := provider.GetToken(context.Background())
	assert.ErrorContains(t, err, "test-error: test-description")
	assert.Empty(t, token)
}

func Test_isTokenValid(t *testing.T) {
	type testcase struct {
		token     *string
		expiresAt *time.Time
		expBuffer time.Duration

		expected bool
	}

	testcases := []testcase{
		{token: nil, expiresAt: nil, expected: false},
		{token: ptr("foo"), expiresAt: nil, expected: false},
		{token: ptr("foo"), expiresAt: ptr(time.Now().Add(-1)), expected: false},
		{token: ptr("foo"), expiresAt: ptr(time.Now().Add(1 * time.Minute)), expected: true},
		{token: ptr("foo"), expiresAt: ptr(time.Now().Add(10 * time.Second)), expBuffer: 9 * time.Second, expected: true},
		{token: ptr("foo"), expiresAt: ptr(time.Now().Add(10 * time.Second)), expBuffer: 10 * time.Second, expected: false},
		{token: ptr("foo"), expiresAt: ptr(time.Now().Add(10 * time.Second)), expBuffer: 11 * time.Second, expected: false},
		{token: ptr("foo"), expiresAt: ptr(time.Now()), expBuffer: 10 * time.Second, expected: false},
	}

	for _, testcase := range testcases {
		t.Run(t.Name(), func(t *testing.T) {
			assert.Equal(t, testcase.expected, isTokenValid(testcase.token, testcase.expiresAt, testcase.expBuffer))
		})
	}
}

func ptr[T any](val T) *T {
	return &val
}
