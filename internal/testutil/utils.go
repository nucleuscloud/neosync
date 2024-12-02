package testutil

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
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

func GetTestLogger(t testing.TB) *slog.Logger {
	testHandler := slog.NewTextHandler(testWriter{t}, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	return slog.New(testHandler)
}

type testWriter struct {
	t testing.TB
}

func (tw testWriter) Write(p []byte) (n int, err error) {
	// removes extra line between log statements
	msg := bytes.TrimSuffix(p, []byte("\n"))
	tw.t.Log(string(msg))
	return len(p), nil
}
