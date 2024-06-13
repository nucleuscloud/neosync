package pgxslog

import (
	"context"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5/tracelog"
	"github.com/spf13/viper"
)

// Somewhat pulled from: https://github.com/mcosta74/pgx-slog
// Copied to avoid adding another go mod dependency which also allows us to ensure this is always compatible with our version of pgx

type Logger struct {
	l        *slog.Logger
	omitArgs bool
}

func NewLogger(l *slog.Logger, omitArgs bool) *Logger {
	return &Logger{l: l, omitArgs: omitArgs}
}

func (l *Logger) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
	if level == tracelog.LogLevelNone {
		return
	}

	attrs := make([]slog.Attr, 0, len(data))
	for k, v := range data {
		if k == "args" && l.omitArgs {
			continue
		}
		attrs = append(attrs, slog.Any(k, v))
	}

	var lvl slog.Level
	switch level {
	case tracelog.LogLevelTrace:
		lvl = slog.LevelDebug - 1
		attrs = append(attrs, slog.Any("PGX_LOG_LEVEL", level))
	case tracelog.LogLevelDebug:
		lvl = slog.LevelDebug
	case tracelog.LogLevelInfo:
		lvl = slog.LevelInfo
	case tracelog.LogLevelWarn:
		lvl = slog.LevelWarn
	case tracelog.LogLevelError:
		lvl = slog.LevelError
	default:
		lvl = slog.LevelError
		attrs = append(attrs, slog.Any("INVALID_PGX_LOG_LEVEL", level))
	}

	attrAny := make([]any, len(attrs))
	for i, attr := range attrs {
		attrAny[i] = attr
	}
	dbGroup := slog.Group("db", attrAny...)

	l.l.LogAttrs(ctx, lvl, msg, dbGroup)
}

// Returns a tracelog.LogLevel as configured by the environment
func GetDatabaseLogLevel() tracelog.LogLevel {
	input := viper.GetString("DB_LOG_LEVEL")
	ll, err := tracelog.LogLevelFromString(strings.ToLower(input))
	if err != nil {
		return tracelog.LogLevelNone
	}
	return ll
}

func GetShouldOmitArgs() bool {
	return viper.GetBool("DB_OMIT_ARGS")
}
