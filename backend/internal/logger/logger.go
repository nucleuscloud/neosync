package neosynclogger

import (
	"log"
	"log/slog"
	"os"

	"github.com/spf13/viper"
)

func New(formatAsJson bool, opts *slog.HandlerOptions) *slog.Logger {
	if formatAsJson {
		return slog.New(slog.NewJSONHandler(os.Stdout, opts))
	}
	return slog.New(slog.NewTextHandler(os.Stdout, opts))
}

func NewLogLogger(formatAsJson bool, opts *slog.HandlerOptions) *log.Logger {
	if formatAsJson {
		return slog.NewLogLogger(slog.NewJSONHandler(os.Stdout, opts), slog.LevelInfo)
	}
	return slog.NewLogLogger(slog.NewTextHandler(os.Stdout, opts), slog.LevelInfo)
}

func ShouldFormatAsJson() bool {
	// using GetString instead of GetBool so that we can default to true if the env var is not present
	result := viper.GetString("LOGS_FORMAT_JSON")

	if result == "" {
		return true
	}
	return result == "true"
}
