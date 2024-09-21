package neomigrate

import (
	"fmt"
	"log/slog"
)

type migrateLogger struct {
	logger  *slog.Logger
	verbose bool
}

func newMigrateLogger(logger *slog.Logger, verbose bool) *migrateLogger {
	return &migrateLogger{logger: logger, verbose: verbose}
}

func (m *migrateLogger) Verbose() bool {
	return m.verbose
}
func (m *migrateLogger) Printf(format string, v ...any) {
	m.logger.Info(fmt.Sprintf("migrate: %s", fmt.Sprintf(format, v...)))
}
