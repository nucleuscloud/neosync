package cli_logger

import (
	"log/slog"
	"os"

	charmlog "github.com/charmbracelet/log"
)

func NewCharmLogger(level charmlog.Level) *charmlog.Logger {
	return charmlog.NewWithOptions(os.Stderr, charmlog.Options{
		ReportTimestamp: true,
		Level:           level,
	})
}

func NewSLogger(level charmlog.Level) *slog.Logger {
	charmlogger := NewCharmLogger(level)
	return slog.New(charmlogger)
}

// Returns both a charm logger and the wrapped slog logger
func NewSlogCharmLogger(level charmlog.Level) (charmlogger *charmlog.Logger, slogger *slog.Logger) {
	charmlogger = NewCharmLogger(level)
	slogger = slog.New(charmlogger)
	return
}

func GetCharmLevelOrDefault(isDebug bool) charmlog.Level {
	if isDebug {
		return charmlog.DebugLevel
	}
	return charmlog.InfoLevel
}
