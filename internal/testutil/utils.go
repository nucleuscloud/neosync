package testutil

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"
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
	// removes extra line between log statements
	msg := strings.TrimSuffix(string(p), "\n")
	tw.t.Log(msg)
	return len(p), nil
}
