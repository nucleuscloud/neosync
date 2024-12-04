package testutil

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/neilotoole/slogt"
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

// not safe for concurrent use
func GetTestLogger(t testing.TB) *slog.Logger {
	f := slogt.Factory(func(w io.Writer) slog.Handler {
		return slog.NewTextHandler(w, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	})

	return slogt.New(t, f)
}

func GetConcurrentTestLogger(t testing.TB) *slog.Logger {
	if testing.Verbose() {
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	}
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
}
