package testutil

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	charmlog "github.com/charmbracelet/log"
)

func ShouldRunIntegrationTest() bool {
	evkey := "INTEGRATION_TESTS_ENABLED"
	shouldRun := os.Getenv(evkey)
	if shouldRun != "1" {
		slog.Warn(fmt.Sprintf("skipping integration tests, set %s=1 to enable", evkey))
		return false
	}
	return true
}

func GetTestSlogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
}

func GetTestCharmLogger() *charmlog.Logger {
	return charmlog.NewWithOptions(io.Discard, charmlog.Options{
		Level: charmlog.DebugLevel,
	})
}
