package logger_utils

import (
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/viper"
)

func NewLoggers() (slogger *slog.Logger, loglogger *log.Logger) {
	handler := NewJsonLogHandler()
	return slog.New(handler), slog.NewLogLogger(handler, getLogLevel())
}

func NewJsonSLogger() *slog.Logger {
	return slog.New(NewJsonLogHandler())
}

func NewJsonLogLogger() *log.Logger {
	return slog.NewLogLogger(NewJsonLogHandler(), getLogLevel())
}

func NewJsonLogHandler() *slog.JSONHandler {
	return slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
}

func getLogLevel() slog.Level {
	input := viper.GetString("LOG_LEVEL")
	switch strings.ToUpper(input) {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
