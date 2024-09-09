package v1alpha1_useraccountservice

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func Test_getAllowedRecordCount(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		actual := getAllowedRecordCount(pgtype.Int8{Int64: 1, Valid: true})
		require.NotNil(t, actual)
		require.Equal(t, uint64(1), *actual)
	})
	t.Run("invalid", func(t *testing.T) {
		actual := getAllowedRecordCount(pgtype.Int8{})
		require.Nil(t, actual)
	})
}

func Test_toUint64(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		actual := toUint64(1)
		require.Equal(t, uint64(1), actual)
	})
	t.Run("zero", func(t *testing.T) {
		actual := toUint64(0)
		require.Equal(t, uint64(0), actual)
	})
	t.Run("negative", func(t *testing.T) {
		actual := toUint64(-1)
		require.Equal(t, uint64(0), actual)
	})
}
