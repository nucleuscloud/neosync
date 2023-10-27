package auth_jwt

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
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
	ctx := context.WithValue(context.Background(), tokenContextKey{}, data)

	ctxdata, err := GetTokenDataFromCtx(ctx)
	assert.Nil(t, err)
	assert.Equal(t, ctxdata, data)
}

func Test_New(t *testing.T) {
	_, err := New(nil)
	assert.Error(t, err)

	_, err = New(&ClientConfig{BaseUrl: "", ApiAudiences: []string{"foo"}})
	assert.Nil(t, err)

	_, err = New(&ClientConfig{BaseUrl: "", ApiAudiences: nil})
	assert.Error(t, err, "fails if api audiences is nil")
}
