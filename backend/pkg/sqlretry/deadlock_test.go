package sqlretry

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
)

func TestIsDeadlockError(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		if isDeadlockError(nil) {
			t.Error("expected false for nil error")
		}
	})

	t.Run("MySQL", func(t *testing.T) {
		t.Run("deadlock error", func(t *testing.T) {
			err := &mysql.MySQLError{Number: mysqlDeadlock}
			if !isDeadlockError(err) {
				t.Error("expected true for MySQL deadlock error")
			}
		})

		t.Run("wrapped deadlock error", func(t *testing.T) {
			originalErr := &mysql.MySQLError{Number: mysqlDeadlock}
			wrappedErr := fmt.Errorf("database error: %w", originalErr)
			doubleWrappedErr := fmt.Errorf("transaction failed: %w", wrappedErr)
			if !isDeadlockError(doubleWrappedErr) {
				t.Error("expected true for wrapped MySQL deadlock error")
			}
		})

		t.Run("lock timeout error", func(t *testing.T) {
			err := &mysql.MySQLError{Number: mysqlLockTimeout}
			if !isDeadlockError(err) {
				t.Error("expected true for MySQL lock timeout error")
			}
		})

		t.Run("wrapped lock timeout error", func(t *testing.T) {
			originalErr := &mysql.MySQLError{Number: mysqlLockTimeout}
			wrappedErr := fmt.Errorf("database error: %w", originalErr)
			if !isDeadlockError(wrappedErr) {
				t.Error("expected true for wrapped MySQL lock timeout error")
			}
		})

		t.Run("non-deadlock error", func(t *testing.T) {
			err := &mysql.MySQLError{Number: 1000}
			if isDeadlockError(err) {
				t.Error("expected false for non-deadlock MySQL error")
			}
		})

		t.Run("wrapped non-deadlock error", func(t *testing.T) {
			originalErr := &mysql.MySQLError{Number: 1000}
			wrappedErr := fmt.Errorf("database error: %w", originalErr)
			if isDeadlockError(wrappedErr) {
				t.Error("expected false for wrapped non-deadlock MySQL error")
			}
		})
	})

	t.Run("PostgreSQL", func(t *testing.T) {
		t.Run("deadlock error", func(t *testing.T) {
			err := &pq.Error{Code: pqDeadlockDetected}
			if !isDeadlockError(err) {
				t.Error("expected true for PostgreSQL deadlock error")
			}
		})

		t.Run("wrapped deadlock error", func(t *testing.T) {
			originalErr := &pq.Error{Code: pqDeadlockDetected}
			wrappedErr := fmt.Errorf("database error: %w", originalErr)
			doubleWrappedErr := fmt.Errorf("transaction failed: %w", wrappedErr)
			if !isDeadlockError(doubleWrappedErr) {
				t.Error("expected true for wrapped PostgreSQL deadlock error")
			}
		})

		t.Run("non-deadlock error", func(t *testing.T) {
			err := &pq.Error{Code: "23505"} // unique violation error
			if isDeadlockError(err) {
				t.Error("expected false for non-deadlock PostgreSQL error")
			}
		})

		t.Run("wrapped non-deadlock error", func(t *testing.T) {
			originalErr := &pq.Error{Code: "23505"} // unique violation error
			wrappedErr := fmt.Errorf("database error: %w", originalErr)
			if isDeadlockError(wrappedErr) {
				t.Error("expected false for wrapped non-deadlock PostgreSQL error")
			}
		})
	})

	t.Run("MSSQL", func(t *testing.T) {
		t.Run("error 1205", func(t *testing.T) {
			err := errors.New("SQL Server Error 1205: Transaction (Process ID 52) was deadlocked")
			if !isDeadlockError(err) {
				t.Error("expected true for MSSQL error 1205")
			}
		})

		t.Run("deadlock victim message", func(t *testing.T) {
			err := errors.New("Transaction was chosen as the deadlock victim")
			if !isDeadlockError(err) {
				t.Error("expected true for MSSQL deadlock victim message")
			}
		})

		t.Run("process ID message", func(t *testing.T) {
			err := errors.New("Transaction (Process ID 64) was involved in a deadlock")
			if !isDeadlockError(err) {
				t.Error("expected true for MSSQL process ID message")
			}
		})

		t.Run("non-deadlock error", func(t *testing.T) {
			err := errors.New("SQL Server Error 1234: General error")
			if isDeadlockError(err) {
				t.Error("expected false for non-deadlock MSSQL error")
			}
		})
	})

	t.Run("generic error", func(t *testing.T) {
		err := errors.New("some random error")
		if isDeadlockError(err) {
			t.Error("expected false for generic error")
		}
	})
}

func TestIsMysqlDeadlockError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "deadlock error",
			err:      &mysql.MySQLError{Number: mysqlDeadlock},
			expected: true,
		},
		{
			name:     "lock timeout error",
			err:      &mysql.MySQLError{Number: mysqlLockTimeout},
			expected: true,
		},
		{
			name:     "non-deadlock MySQL error",
			err:      &mysql.MySQLError{Number: 1000},
			expected: false,
		},
		{
			name:     "non-MySQL error",
			err:      errors.New("generic error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isMysqlDeadlockError(tt.err); got != tt.expected {
				t.Errorf("isMysqlDeadlockError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsPostgresDeadlock(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "deadlock error (lib/pq)",
			err:      &pq.Error{Code: pqDeadlockDetected},
			expected: true,
		},
		{
			name:     "deadlock error (pgx)",
			err:      &pgconn.PgError{Code: pqDeadlockDetected},
			expected: true,
		},
		{
			name:     "non-deadlock postgres error (lib/pq)",
			err:      &pq.Error{Code: "23505"},
			expected: false,
		},
		{
			name:     "non-deadlock postgres error (pgx)",
			err:      &pgconn.PgError{Code: "23505"},
			expected: false,
		},
		{
			name:     "non-postgres error",
			err:      errors.New("generic error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPostgresDeadlock(tt.err); got != tt.expected {
				t.Errorf("isPostgresDeadlock() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsMSSQLDeadlock(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "error 1205",
			err:      errors.New("SQL Server Error 1205: Transaction was deadlocked"),
			expected: true,
		},
		{
			name:     "deadlock victim message",
			err:      errors.New("Transaction was chosen as the deadlock victim"),
			expected: true,
		},
		{
			name:     "process ID message",
			err:      errors.New("Transaction (Process ID 64) was deadlocked"),
			expected: true,
		},
		{
			name:     "case insensitive check",
			err:      errors.New("TRANSACTION WAS CHOSEN AS THE DEADLOCK VICTIM"),
			expected: true,
		},
		{
			name:     "non-deadlock error",
			err:      errors.New("SQL Server Error 1234: General error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isMSSQLDeadlock(tt.err); got != tt.expected {
				t.Errorf("isMSSQLDeadlock() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func Test_Nick(t *testing.T) {
	err := &pq.Error{Code: pqDeadlockDetected}
	t.Log(err.Error())
}
