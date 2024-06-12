package pgxslog

import (
	"context"
	"log/slog"
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/spf13/viper"
)

// Pulled from: https://github.com/mcosta74/pgx-slog
// Copied to avoid adding another go mod dependency which also allows us to ensure this is always compatible with our version of pgx

type Logger struct {
	l *slog.Logger
}

func NewLogger(l *slog.Logger) *Logger {
	return &Logger{l: l}
}

func (l *Logger) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
	attrs := make([]slog.Attr, 0, len(data))
	for k, v := range data {
		attrs = append(attrs, slog.Any(k, v))
	}

	var lvl slog.Level
	switch level {
	case pgx.LogLevelTrace:
		lvl = slog.LevelDebug - 1
		attrs = append(attrs, slog.Any("PGX_LOG_LEVEL", level))
	case pgx.LogLevelDebug:
		lvl = slog.LevelDebug
	case pgx.LogLevelInfo:
		lvl = slog.LevelInfo
	case pgx.LogLevelWarn:
		lvl = slog.LevelWarn
	case pgx.LogLevelError:
		lvl = slog.LevelError
	default:
		lvl = slog.LevelError
		attrs = append(attrs, slog.Any("INVALID_PGX_LOG_LEVEL", level))
	}
	l.l.LogAttrs(ctx, lvl, msg, attrs...)
}

// Returns a tracelog.LogLevel as configured by the environment
func GetDatabaseLogLevel() tracelog.LogLevel {
	input := viper.GetString("DB_LOG_LEVEL")
	switch strings.ToUpper(input) {
	case "TRACE":
		return tracelog.LogLevelTrace
	case "DEBUG":
		return tracelog.LogLevelDebug
	case "INFO":
		return tracelog.LogLevelInfo
	case "WARN":
		return tracelog.LogLevelWarn
	case "ERROR":
		return tracelog.LogLevelError
	default:
		return tracelog.LogLevelDebug
	}
}
