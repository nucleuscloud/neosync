package neosyncdb

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func Test_IsConflict(t *testing.T) {
	assert.False(t, IsConflict(nil))
	assert.False(t, IsConflict(errors.New("test error")))
	assert.False(t, IsConflict(&pgconn.PgError{}))
	assert.True(t, IsConflict(&pgconn.PgError{Code: PqUniqueViolationCode}))
}

func Test_IsNoRows(t *testing.T) {
	assert.False(t, IsNoRows(nil))
	assert.False(t, IsNoRows(errors.New("test error")))
	assert.True(t, IsNoRows(sql.ErrNoRows))
	assert.True(t, IsNoRows(pgx.ErrNoRows))
	assert.True(t, IsNoRows(fmt.Errorf("test is no rows: %w", pgx.ErrNoRows)), "should work for wrapped errors")
	assert.True(t, IsNoRows(fmt.Errorf("test is no rows: %w", sql.ErrNoRows)), "should work for wrapped errors")
}

func Test_IsTxDone(t *testing.T) {
	assert.False(t, isTxDone(nil))
	assert.True(t, isTxDone(pgx.ErrTxClosed))
	assert.True(t, isTxDone(sql.ErrTxDone))
	assert.True(t, isTxDone(fmt.Errorf("test tx has completed: %w", sql.ErrTxDone)), "should work for wrapped errors")
	assert.True(t, isTxDone(fmt.Errorf("test tx has completed: %w", pgx.ErrTxClosed)), "should work for wrapped errors")
}

func Test_GetDbUrl(t *testing.T) {
	assert.Equal(t, GetDbUrl(nil), "")
	assert.Equal(
		t,
		GetDbUrl(&ConnectConfig{}),
		"postgres://:@:0/",
	)
	sslmode := "disable"
	assert.Equal(
		t,
		GetDbUrl(&ConnectConfig{
			User:     "myuser",
			Pass:     "mypass",
			Host:     "localhost",
			Port:     5432,
			Database: "neosync",
			SslMode:  &sslmode,
		}),
		"postgres://myuser:mypass@localhost:5432/neosync?sslmode=disable",
	)

	migrationsTableName := "test-table-name"
	migrationsTableQuoted := true
	assert.Equal(
		t,
		GetDbUrl(&ConnectConfig{
			User:                  "myuser",
			Pass:                  "mypass",
			Host:                  "localhost",
			Port:                  5432,
			Database:              "neosync",
			SslMode:               &sslmode,
			MigrationsTableName:   &migrationsTableName,
			MigrationsTableQuoted: &migrationsTableQuoted,
		}),
		"postgres://myuser:mypass@localhost:5432/neosync?sslmode=disable&x-migrations-table=test-table-name&x-migrations-table-quoted=true",
	)
}

func Test_ToUuid(t *testing.T) {
	testuuid := uuid.New()

	pguuid, err := ToUuid(testuuid.String())
	assert.Nil(t, err)

	pguuid2 := pgtype.UUID{}
	err = pguuid2.Scan(testuuid.String())
	assert.Nil(t, err)
	assert.Equal(t, pguuid, pguuid2)
}

func Test_UUIDString(t *testing.T) {
	testuuid := uuid.New()

	pguuid := pgtype.UUID{}
	err := pguuid.Scan(testuuid.String())
	assert.Nil(t, err)

	testuuidstr := UUIDString(pguuid)
	assert.Equal(t, testuuidstr, testuuid.String())
}

func Test_UUIDStrings(t *testing.T) {
	testuuid1 := uuid.New()
	testuuid2 := uuid.New()

	pgtestuuid1, err := ToUuid(testuuid1.String())
	assert.Nil(t, err)
	pgtestuuid2, err := ToUuid(testuuid2.String())
	assert.Nil(t, err)

	uuidstrs := UUIDStrings([]pgtype.UUID{pgtestuuid1, pgtestuuid2})

	assert.Equal(
		t,
		uuidstrs,
		[]string{testuuid1.String(), testuuid2.String()},
	)
}

func Test_ToNullableString(t *testing.T) {
	assert.Nil(t, ToNullableString(pgtype.Text{}))

	text := pgtype.Text{}
	err := text.Scan("hello world")
	assert.Nil(t, err)

	output := ToNullableString(text)
	assert.NotNil(t, output)
	assert.Equal(t, *output, "hello world")
}

func Test_Int16ToBool(t *testing.T) {
	assert.False(t, Int16ToBool(0))
	assert.True(t, Int16ToBool(1))
}

func Test_BoolToInt16(t *testing.T) {
	assert.Equal(t, BoolToInt16(false), int16(0))
	assert.Equal(t, BoolToInt16(true), int16(1))
}
