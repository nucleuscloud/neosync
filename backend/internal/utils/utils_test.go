package utils

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_FilterSlice(t *testing.T) {
	assert.Empty(t, FilterSlice[string]([]string{"foo", "bar"}, func(s string) bool { return false }))
	assert.Equal(
		t,
		FilterSlice[string]([]string{"foo", "bar"}, func(s string) bool { return true }),
		[]string{"foo", "bar"},
	)
	assert.Equal(
		t,
		FilterSlice[string]([]string{"foo", "bar"}, func(s string) bool { return s == "foo" }),
		[]string{"foo"},
	)
}

func Test_MapSlice(t *testing.T) {
	assert.Equal(
		t,
		MapSlice[string, string]([]string{"foo", "bar"}, func(s string) string { return fmt.Sprintf("%s_test", s) }),
		[]string{"foo_test", "bar_test"},
	)
	assert.Equal(
		t,
		MapSlice[string, bool]([]string{"foo", "bar"}, func(s string) bool { return true }),
		[]bool{true, true},
	)
}

func Test_ToSha256(t *testing.T) {
	assert.Equal(
		t,
		ToSha256("foobar"),
		"c3ab8ff13720e8ad9047dd39466b3c8974e592c2fa383d4a3960714caef0c4f2",
	)
}

func Test_GetBearerTokenFromHeader(t *testing.T) {
	_, err := GetBearerTokenFromHeader(http.Header{}, "Authorization")
	assert.Error(t, err)
	_, err = GetBearerTokenFromHeader(http.Header{"Authorization": []string{}}, "Authorization")
	assert.Error(t, err)
	_, err = GetBearerTokenFromHeader(http.Header{"Authorization": []string{"Foo"}}, "Authorization")
	assert.Error(t, err)
	_, err = GetBearerTokenFromHeader(http.Header{"Authorization": []string{"Foo Foo Foo"}}, "Authorization")
	assert.Error(t, err)
	_, err = GetBearerTokenFromHeader(http.Header{"Authorization": []string{"Foo Foo"}}, "Authorization")
	assert.Error(t, err)
	_, err = GetBearerTokenFromHeader(http.Header{"Authorization": []string{"Bearer"}}, "Authorization")
	assert.Error(t, err)
	_, err = GetBearerTokenFromHeader(http.Header{"Authorization": []string{"Bearer 123"}}, "Authorizationn")
	assert.Error(t, err)

	token, err := GetBearerTokenFromHeader(http.Header{"Authorization": []string{"Bearer 123"}}, "Authorization")
	assert.Nil(t, err)
	assert.Equal(t, token, "123")
}
