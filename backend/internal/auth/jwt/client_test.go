package auth_jwt

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/auth0/go-jwt-middleware/v2/validator"
)

func Test_hasScope(t *testing.T) {
	assert.True(
		t,
		hasScope([]string{"foo", "bar"}, "foo"),
	)
	assert.False(
		t,
		hasScope([]string{"foo", "bar"}, "fooo"),
	)
}

func Test_TokenContextData_HasScope(t *testing.T) {
	data := &TokenContextData{
		Scopes: []string{"foo", "bar"},
	}
	assert.True(
		t,
		data.HasScope("foo"),
	)
	assert.False(
		t,
		data.HasScope("fooo"),
	)
}

func Test_getCombinedScopesAndPermissions(t *testing.T) {
	assert.Equal(
		t,
		getCombinedScopesAndPermissions("foo bar baz", []string{"foo", "bazz"}),
		[]string{"foo", "bar", "baz", "bazz"},
	)
}

func Test_GetTokenDataFromCtx_Unauthenticated(t *testing.T) {
	data, err := GetTokenDataFromCtx(context.Background())
	assert.Error(t, err)
	assert.Nil(t, data)
}

func Test_GetTokenDataFromCtx_Authenticated(t *testing.T) {
	data := &TokenContextData{}
	ctx := context.WithValue(context.Background(), TokenContextKey{}, data)

	ctxdata, err := GetTokenDataFromCtx(ctx)
	assert.Nil(t, err)
	assert.Equal(t, ctxdata, data)
}

func Test_New(t *testing.T) {
	_, err := New(nil)
	assert.Error(t, err)

	_, err = New(&ClientConfig{BackendIssuerUrl: "", ApiAudiences: []string{"foo"}})
	assert.Nil(t, err)

	_, err = New(&ClientConfig{BackendIssuerUrl: "", ApiAudiences: nil})
	assert.Error(t, err, "fails if api audiences is nil")
}

func Test_Client_InjectTokenCtx(t *testing.T) {
	customclaims := &CustomClaims{
		Scope: "foo bar",
	}
	validatedClaims := &validator.ValidatedClaims{
		RegisteredClaims: validator.RegisteredClaims{
			Subject: "test user",
		},
		CustomClaims: customclaims,
	}
	jwtValidator := &MockJwtValidator{}
	jwtValidator.On("ValidateToken", mock.Anything, "123").Return(validatedClaims, nil)
	client := &Client{jwtValidator: jwtValidator}

	newCtx, err := client.InjectTokenCtx(context.Background(), http.Header{"Authorization": []string{"Bearer 123"}})
	assert.Nil(t, err)

	data, err := GetTokenDataFromCtx(newCtx)
	assert.Nil(t, err)
	assert.Equal(
		t,
		data,
		&TokenContextData{
			ParsedToken: validatedClaims,
			RawToken:    "123",
			Claims:      customclaims,
			AuthUserId:  "test user",
			Scopes:      []string{"foo", "bar"},
		},
	)
}

func Test_Client_InjectTokenCtx_InvalidHeader(t *testing.T) {
	client := &Client{}
	_, err := client.InjectTokenCtx(context.Background(), http.Header{"Authorization": []string{}})
	assert.Error(t, err)
}

func Test_Client_InjectTokenCtx_InvalidToken(t *testing.T) {
	jwtValidator := &MockJwtValidator{}
	jwtValidator.On("ValidateToken", mock.Anything, "123").Return(nil, errors.New("invalid token"))
	client := &Client{jwtValidator: jwtValidator}

	_, err := client.InjectTokenCtx(context.Background(), http.Header{"Authorization": []string{"Bearer 123"}})
	assert.Error(t, err)
}

func Test_Client_InjectTokenCtx_InvalidTokenClaims(t *testing.T) {
	validatedClaims := map[string]string{}
	jwtValidator := &MockJwtValidator{}
	jwtValidator.On("ValidateToken", mock.Anything, "123").Return(validatedClaims, nil)
	client := &Client{jwtValidator: jwtValidator}

	_, err := client.InjectTokenCtx(context.Background(), http.Header{"Authorization": []string{"Bearer 123"}})
	assert.Error(t, err)
}

func Test_Client_InjectTokenCtx_InvalidClaims(t *testing.T) {
	validatedClaims := &validator.ValidatedClaims{
		RegisteredClaims: validator.RegisteredClaims{
			Subject: "test user",
		},
		CustomClaims: nil,
	}
	jwtValidator := &MockJwtValidator{}
	jwtValidator.On("ValidateToken", mock.Anything, "123").Return(validatedClaims, nil)
	client := &Client{jwtValidator: jwtValidator}

	_, err := client.InjectTokenCtx(context.Background(), http.Header{"Authorization": []string{"Bearer 123"}})
	assert.Error(t, err)
}
