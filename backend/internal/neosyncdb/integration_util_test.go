package neosyncdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func requireNoErrResp[T any](t testing.TB, resp T, err error) {
	t.Helper()
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func requireErrResp[T any](t testing.TB, resp T, err error) {
	t.Helper()
	require.Error(t, err)
	require.Nil(t, resp)
}

func assertNoErrResp[T any](t testing.TB, resp T, err error) {
	t.Helper()
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}
