package ee_slack

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	sym_encrypt "github.com/nucleuscloud/neosync/internal/encrypt/sym"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_NewClient(t *testing.T) {
	t.Run("creates client successfully with options", func(t *testing.T) {
		encryptor := sym_encrypt.NewMockInterface(t)
		client := NewClient(
			encryptor,
			WithAuthClientCreds("test-client-id", "test-client-secret"),
			WithScope("test-scope"),
			WithRedirectUrl("http://test.com"),
		)
		assert.NotNil(t, client)
	})
}

func Test_Client_GetAuthorizeUrl(t *testing.T) {
	t.Run("gets authorize url successfully", func(t *testing.T) {
		encryptor := sym_encrypt.NewMockInterface(t)
		client := NewClient(
			encryptor,
			WithAuthClientCreds("test-client-id", "test-client-secret"),
			WithScope("test-scope"),
			WithRedirectUrl("http://test.com"),
		)

		encryptor.EXPECT().Encrypt(mock.Anything).Return("encrypted-token", nil).Once()

		authorizedUrl, err := client.GetAuthorizeUrl("test-account-id", "test-user-id")
		assert.NoError(t, err)
		assert.Equal(t, "https://slack.com/oauth/v2/authorize?client_id=test-client-id&redirect_uri=http%3A%2F%2Ftest.com&scope=test-scope&state=encrypted-token", authorizedUrl)
		encryptor.AssertExpectations(t)
	})
}

func Test_Client_ValidateState(t *testing.T) {
	t.Run("validates state successfully", func(t *testing.T) {
		encryptor := sym_encrypt.NewMockInterface(t)
		client := NewClient(
			encryptor,
			WithAuthClientCreds("test-client-id", "test-client-secret"),
			WithScope("test-scope"),
			WithRedirectUrl("http://test.com"),
		)

		state := OauthState{
			AccountId: "test-account-id",
			UserId:    "test-user-id",
			Timestamp: time.Now().Unix(),
		}

		stateJson, err := json.Marshal(state)
		assert.NoError(t, err)

		encryptor.EXPECT().Decrypt("encrypted-token").Return(string(stateJson), nil).Once()

		oauthState, err := client.ValidateState(context.Background(), "encrypted-token", "test-user-id", func(ctx context.Context, userId, accountId string) (bool, error) {
			return true, nil
		})
		assert.NoError(t, err)
		assert.Equal(t, state, *oauthState)
		encryptor.AssertExpectations(t)
	})

	t.Run("invalid account id", func(t *testing.T) {
		encryptor := sym_encrypt.NewMockInterface(t)
		client := NewClient(
			encryptor,
			WithAuthClientCreds("test-client-id", "test-client-secret"),
			WithScope("test-scope"),
			WithRedirectUrl("http://test.com"),
		)

		state := OauthState{
			AccountId: "test-account-id",
			UserId:    "test-user-id",
			Timestamp: time.Now().Unix(),
		}

		stateJson, err := json.Marshal(state)
		assert.NoError(t, err)

		encryptor.EXPECT().Decrypt("encrypted-token").Return(string(stateJson), nil).Once()

		oauthState, err := client.ValidateState(context.Background(), "encrypted-token", "test-user-id", func(ctx context.Context, userId, accountId string) (bool, error) {
			return false, nil
		})
		assert.Error(t, err)
		assert.Nil(t, oauthState)
		encryptor.AssertExpectations(t)
	})

	t.Run("invalid user id", func(t *testing.T) {
		encryptor := sym_encrypt.NewMockInterface(t)
		client := NewClient(
			encryptor,
			WithAuthClientCreds("test-client-id", "test-client-secret"),
			WithScope("test-scope"),
			WithRedirectUrl("http://test.com"),
		)

		state := OauthState{
			AccountId: "test-account-id",
			UserId:    "test-user-id",
			Timestamp: time.Now().Unix(),
		}

		stateJson, err := json.Marshal(state)
		assert.NoError(t, err)

		encryptor.EXPECT().Decrypt("encrypted-token").Return(string(stateJson), nil).Once()

		oauthState, err := client.ValidateState(context.Background(), "encrypted-token", "invalid-user-id", func(ctx context.Context, userId, accountId string) (bool, error) {
			return true, nil
		})
		assert.Error(t, err)
		assert.Nil(t, oauthState)
		encryptor.AssertExpectations(t)
	})

	t.Run("oauth state expired", func(t *testing.T) {
		encryptor := sym_encrypt.NewMockInterface(t)
		client := NewClient(
			encryptor,
			WithAuthClientCreds("test-client-id", "test-client-secret"),
			WithScope("test-scope"),
			WithRedirectUrl("http://test.com"),
		)

		state := OauthState{
			AccountId: "test-account-id",
			UserId:    "test-user-id",
			Timestamp: time.Now().Unix() - 901,
		}

		stateJson, err := json.Marshal(state)
		assert.NoError(t, err)

		encryptor.EXPECT().Decrypt("encrypted-token").Return(string(stateJson), nil).Once()

		oauthState, err := client.ValidateState(context.Background(), "encrypted-token", "test-user-id", func(ctx context.Context, userId, accountId string) (bool, error) {
			return true, nil
		})
		assert.Error(t, err)
		assert.Nil(t, oauthState)
		encryptor.AssertExpectations(t)
	})
}
