package neomigrate

import (
	"context"
	"errors"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Down(
	ctx context.Context,
	connectionString string,
	migrationDirectory string,
	logger *slog.Logger,
) error {
	sourceUrl, err := getMigrationSourceUrl(migrationDirectory)
	if err != nil {
		return err
	}

	m, err := migrate.New(
		sourceUrl,
		connectionString,
	)
	if err != nil {
		return err
	}
	m.Log = newMigrateLogger(logger, false)
	err = m.Down()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	sourcerr, dberr := m.Close()
	if sourcerr != nil {
		return sourcerr
	}
	if dberr != nil {
		return dberr
	}
	return nil
}
