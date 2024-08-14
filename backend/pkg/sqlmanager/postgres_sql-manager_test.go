package sqlmanager

import (
	context "context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func (s *PostgresIntegrationTestSuite) Test_NewPooledSqlDb() {
	t := s.T()

	conn, err := s.sqlmanager.NewPooledSqlDb(s.ctx, slog.Default(), s.mgmtconn)
	requireNoConnErr(t, conn, err)
	requireValidDatabase(t, s.ctx, conn, "postgres", "SELECT 1")
	conn.Db.Close()
}

func (s *PostgresIntegrationTestSuite) Test_NewSqlDb() {
	t := s.T()

	connTimeout := 5
	conn, err := s.sqlmanager.NewSqlDb(s.ctx, slog.Default(), s.mgmtconn, &connTimeout)
	requireNoConnErr(t, conn, err)

	requireValidDatabase(t, s.ctx, conn, "postgres", "SELECT 1")
	conn.Db.Close()
}

func (s *PostgresIntegrationTestSuite) Test_NewSqlDbFromUrl() {
	t := s.T()
	conn, err := s.sqlmanager.NewSqlDbFromUrl(s.ctx, "postgres", s.pgcfg.GetUrl())
	requireNoConnErr(t, conn, err)

	requireValidDatabase(t, s.ctx, conn, "postgres", "SELECT 1")
	conn.Db.Close()
}

func (s *PostgresIntegrationTestSuite) Test_NewSqlDbFromConnectionConfig() {
	t := s.T()
	connTimeout := 5
	conn, err := s.sqlmanager.NewSqlDbFromConnectionConfig(s.ctx, slog.Default(), s.mgmtconn.GetConnectionConfig(), &connTimeout)
	requireNoConnErr(t, conn, err)

	requireValidDatabase(t, s.ctx, conn, "postgres", "SELECT 1")
	conn.Db.Close()
}

func requireNoConnErr(t testing.TB, conn *SqlConnection, err error) {
	require.NoError(t, err)
	require.NotNil(t, conn)
}

func requireValidDatabase(t testing.TB, ctx context.Context, conn *SqlConnection, driver, statement string) {
	require.Equal(t, conn.Driver, driver)
	err := conn.Db.Exec(ctx, statement)
	require.NoError(t, err)
}
