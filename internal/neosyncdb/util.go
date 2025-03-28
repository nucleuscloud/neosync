package neosyncdb

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lib/pq"
)

const (
	PqUniqueViolationCode = "23505"
)

func IsConflict(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == PqUniqueViolationCode {
		return true
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == PqUniqueViolationCode {
		return true
	}

	return false
}

func IsNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows)
}

func isTxDone(err error) bool {
	return errors.Is(err, pgx.ErrTxClosed) || errors.Is(err, sql.ErrTxDone)
}

func GetDbUrl(cfg *ConnectConfig) string {
	if cfg == nil {
		return ""
	}
	dburl := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		cfg.User,
		cfg.Pass,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)
	pgOpts := url.Values{}
	if cfg.SslMode != nil && *cfg.SslMode != "" {
		pgOpts["sslmode"] = []string{*cfg.SslMode}
	}
	if cfg.MigrationsTableName != nil && *cfg.MigrationsTableName != "" {
		pgOpts["x-migrations-table"] = []string{*cfg.MigrationsTableName}
	}
	if cfg.MigrationsTableQuoted != nil {
		pgOpts["x-migrations-table-quoted"] = []string{
			strconv.FormatBool(*cfg.MigrationsTableQuoted),
		}
	}
	if cfg.Options != nil {
		pgOpts["options"] = []string{*cfg.Options}
	}
	if len(pgOpts) > 0 {
		return fmt.Sprintf("%s?%s", dburl, pgOpts.Encode())
	}
	return dburl
}

func UUIDString(value pgtype.UUID) string {
	return fmt.Sprintf(
		"%x-%x-%x-%x-%x",
		value.Bytes[0:4],
		value.Bytes[4:6],
		value.Bytes[6:8],
		value.Bytes[8:10],
		value.Bytes[10:16],
	)
}

func UUIDStrings(values []pgtype.UUID) []string {
	outputs := []string{}
	for _, value := range values {
		outputs = append(outputs, UUIDString(value))
	}
	return outputs
}

func ToUuid(value string) (pgtype.UUID, error) {
	uuid := pgtype.UUID{}
	err := uuid.Scan(value)
	return uuid, err
}

func ToTimestamp(value time.Time) (pgtype.Timestamp, error) {
	timestamp := pgtype.Timestamp{}
	err := timestamp.Scan(value)
	return timestamp, err
}

func ToNullableString(text pgtype.Text) *string {
	if text.Valid {
		return &text.String
	}
	return nil
}

func Int16ToBool(val int16) bool {
	return val > 0
}

func BoolToInt16(val bool) int16 {
	if val {
		return 1
	}
	return 0
}
