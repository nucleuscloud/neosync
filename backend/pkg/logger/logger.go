package neosynclogger

import (
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/viper"
	slogdedup "github.com/veqryn/slog-dedup"
)

func NewLoggers() (slogger *slog.Logger, loglogger *log.Logger) {
	handler := newHandler()
	return slog.New(handler), slog.NewLogLogger(handler, getLogLevel())
}

func NewJsonSLogger() *slog.Logger {
	return slog.New(newJsonLogHandler())
}

func NewJsonLogLogger() *log.Logger {
	return slog.NewLogLogger(newJsonLogHandler(), getLogLevel())
}

// Returns either JSON or TEXT depending on the environment variable LOGS_FORMAT_JSON
// Defaults to JSON
func newHandler() slog.Handler {
	if ShouldFormatAsJson() {
		return newJsonLogHandler()
	}
	return newTextLogHandler()
}

func newJsonLogHandler() slog.Handler {
	jsonhandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: getLogLevel(),
	})
	return slogdedup.NewOverwriteHandler(jsonhandler, nil)
}

func newTextLogHandler() *slog.TextHandler {
	return slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: getLogLevel(),
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

func ShouldFormatAsJson() bool {
	// using GetString instead of GetBool so that we can default to true if the env var is not present
	result := viper.GetString("LOGS_FORMAT_JSON")

	if result == "" {
		return true
	}
	return result == "true"
}
