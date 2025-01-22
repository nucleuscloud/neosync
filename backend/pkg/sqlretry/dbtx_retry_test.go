package sqlretry

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	"github.com/stretchr/testify/mock"
)

func TestIsRetryableError(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		if isRetryableError(nil) {
			t.Error("expected false for nil error")
		}
	})

	t.Run("MySQL", func(t *testing.T) {
		t.Run("deadlock error", func(t *testing.T) {
			err := &mysql.MySQLError{Number: mysqlDeadlock}
			if !isRetryableError(err) {
				t.Error("expected true for MySQL deadlock error")
			}
		})

		t.Run("lock timeout error", func(t *testing.T) {
			err := &mysql.MySQLError{Number: mysqlLockTimeout}
			if !isRetryableError(err) {
				t.Error("expected true for MySQL lock timeout error")
			}
		})

		t.Run("busy buffer error", func(t *testing.T) {
			err := mysql.ErrBusyBuffer
			if !isRetryableError(err) {
				t.Error("expected true for MySQL busy buffer error")
			}
		})

		t.Run("non-retryable error", func(t *testing.T) {
			err := &mysql.MySQLError{Number: 1064} // syntax error
			if isRetryableError(err) {
				t.Error("expected false for non-retryable MySQL error")
			}
		})
	})

	t.Run("PostgreSQL", func(t *testing.T) {
		t.Run("deadlock error", func(t *testing.T) {
			err := &pq.Error{Code: pqDeadlockDetected}
			if !isRetryableError(err) {
				t.Error("expected true for PostgreSQL deadlock error")
			}
		})

		t.Run("serialization failure", func(t *testing.T) {
			err := &pq.Error{Code: pqSerializationFailure}
			if !isRetryableError(err) {
				t.Error("expected true for PostgreSQL serialization failure")
			}
		})

		t.Run("lock not available", func(t *testing.T) {
			err := &pq.Error{Code: pqLockNotAvailable}
			if !isRetryableError(err) {
				t.Error("expected true for PostgreSQL lock not available")
			}
		})

		t.Run("object in use", func(t *testing.T) {
			err := &pq.Error{Code: pqObjectInUse}
			if !isRetryableError(err) {
				t.Error("expected true for PostgreSQL object in use")
			}
		})

		t.Run("non-retryable error", func(t *testing.T) {
			err := &pq.Error{Code: "42601"} // syntax error
			if isRetryableError(err) {
				t.Error("expected false for non-retryable PostgreSQL error")
			}
		})
	})

	t.Run("MSSQL", func(t *testing.T) {
		t.Run("deadlock error", func(t *testing.T) {
			err := errors.New("Transaction (Process ID 52) was deadlocked")
			if !isRetryableError(err) {
				t.Error("expected true for MSSQL deadlock error")
			}
		})

		t.Run("deadlock victim", func(t *testing.T) {
			err := errors.New("Transaction was chosen as the deadlock victim")
			if !isRetryableError(err) {
				t.Error("expected true for MSSQL deadlock victim")
			}
		})

		t.Run("non-retryable error", func(t *testing.T) {
			err := errors.New("Incorrect syntax near")
			if isRetryableError(err) {
				t.Error("expected false for non-retryable MSSQL error")
			}
		})
	})

	t.Run("non-retryable generic error", func(t *testing.T) {
		err := errors.New("some random error")
		if isRetryableError(err) {
			t.Error("expected false for non-retryable generic error")
		}
	})
}

func TestRetryDBTX(t *testing.T) {
	t.Run("ExecContext", func(t *testing.T) {
		t.Run("succeeds first try", func(t *testing.T) {
			mockDB := mysql_queries.NewMockDBTX(t)
			expectedResult := &mockResult{}

			mockDB.EXPECT().
				ExecContext(mock.Anything, "SELECT 1", mock.Anything).
				Return(expectedResult, nil).
				Once()

			db := New(mockDB)
			result, err := db.ExecContext(context.Background(), "SELECT 1", 1)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != expectedResult {
				t.Error("expected result to match")
			}
		})

		t.Run("retries on deadlock and succeeds", func(t *testing.T) {
			mockDB := mysql_queries.NewMockDBTX(t)
			expectedResult := &mockResult{}

			mockDB.EXPECT().
				ExecContext(mock.Anything, "SELECT 1", mock.Anything).
				Return(nil, &mysql.MySQLError{Number: mysqlDeadlock}).
				Once()

			mockDB.EXPECT().
				ExecContext(mock.Anything, "SELECT 1", mock.Anything).
				Return(expectedResult, nil).
				Once()

			db := New(mockDB)
			result, err := db.ExecContext(context.Background(), "SELECT 1", 1)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != expectedResult {
				t.Error("expected result to match")
			}
		})

		t.Run("does not retry on non-retryable error", func(t *testing.T) {
			mockDB := mysql_queries.NewMockDBTX(t)
			expectedErr := errors.New("non-retryable error")

			mockDB.EXPECT().
				ExecContext(mock.Anything, "SELECT 1", mock.Anything).
				Return(nil, expectedErr).
				Once()

			db := New(mockDB)
			_, err := db.ExecContext(context.Background(), "SELECT 1", 1)

			if err != expectedErr {
				t.Errorf("expected error %v, got %v", expectedErr, err)
			}
		})
	})

	t.Run("QueryContext", func(t *testing.T) {
		t.Run("succeeds first try", func(t *testing.T) {
			mockDB := mysql_queries.NewMockDBTX(t)
			expectedRows := &sql.Rows{}

			mockDB.EXPECT().
				QueryContext(mock.Anything, "SELECT 1", mock.Anything).
				Return(expectedRows, nil).
				Once()

			db := New(mockDB)
			rows, err := db.QueryContext(context.Background(), "SELECT 1", 1)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if rows != expectedRows {
				t.Error("expected rows to match")
			}
		})

		t.Run("retries on deadlock and succeeds", func(t *testing.T) {
			mockDB := mysql_queries.NewMockDBTX(t)
			expectedRows := &sql.Rows{}

			mockDB.EXPECT().
				QueryContext(mock.Anything, "SELECT 1", mock.Anything).
				Return(nil, &pq.Error{Code: pqDeadlockDetected}).
				Once()

			mockDB.EXPECT().
				QueryContext(mock.Anything, "SELECT 1", mock.Anything).
				Return(expectedRows, nil).
				Once()

			db := New(mockDB)
			rows, err := db.QueryContext(context.Background(), "SELECT 1", 1)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if rows != expectedRows {
				t.Error("expected rows to match")
			}
		})
	})

	t.Run("PrepareContext", func(t *testing.T) {
		t.Run("succeeds first try", func(t *testing.T) {
			mockDB := mysql_queries.NewMockDBTX(t)
			expectedStmt := &sql.Stmt{}

			mockDB.EXPECT().
				PrepareContext(mock.Anything, "SELECT 1").
				Return(expectedStmt, nil).
				Once()

			db := New(mockDB)
			stmt, err := db.PrepareContext(context.Background(), "SELECT 1")

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if stmt != expectedStmt {
				t.Error("expected statement to match")
			}
		})
	})

	t.Run("QueryRowContext", func(t *testing.T) {
		t.Run("passes through to underlying db", func(t *testing.T) {
			mockDB := mysql_queries.NewMockDBTX(t)
			expectedRow := &sql.Row{}

			mockDB.EXPECT().
				QueryRowContext(mock.Anything, "SELECT 1", mock.Anything).
				Return(expectedRow).
				Once()

			db := New(mockDB)
			row := db.QueryRowContext(context.Background(), "SELECT 1", 1)

			if row != expectedRow {
				t.Error("expected row to match")
			}
		})
	})
}

// mockResult implements sql.Result for testing
type mockResult struct{}

func (m *mockResult) LastInsertId() (int64, error) { return 0, nil }
func (m *mockResult) RowsAffected() (int64, error) { return 0, nil }
