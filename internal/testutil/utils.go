package testutil

import (
	"fmt"
	"log/slog"
	"os"
	"testing"

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

func GetTestLogger(t *testing.T) *slog.Logger {
	testHandler := slog.NewTextHandler(testWriter{t}, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	return slog.New(testHandler)
}

type testWriter struct {
	t *testing.T
}

func (tw testWriter) Write(p []byte) (n int, err error) {
	tw.t.Log(string(p))
	return len(p), nil
}

func GetTestCharmSlogger() *slog.Logger {
	charmlogger := charmlog.NewWithOptions(os.Stdout, charmlog.Options{
		Level: charmlog.DebugLevel,
	})
	return slog.New(charmlogger)
}
