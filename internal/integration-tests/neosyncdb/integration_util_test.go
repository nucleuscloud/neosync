package neosyncdb

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nucleuscloud/neosync/internal/neosyncdb"
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

func getFutureTs(t testing.TB, d time.Duration) pgtype.Timestamp {
	t.Helper()
	ts, err := neosyncdb.ToTimestamp(time.Now().Add(d))
	require.NoError(t, err)
	return ts
}
