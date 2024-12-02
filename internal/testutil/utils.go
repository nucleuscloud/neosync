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

func ShouldRunS3IntegrationTest() bool {
	evkey := "S3_INTEGRATION_TESTS_ENABLED"
	shouldRun := os.Getenv(evkey)
	if shouldRun != "1" {
		slog.Warn(fmt.Sprintf("skipping S3 integration tests, set %s=1 to enable", evkey))
		return false
	}
	return true
}

type AwsS3Config struct {
	Bucket          string
	Region          string
	AccessKeyId     string
	SecretAccessKey string
}

func GetTestAwsS3Config() *AwsS3Config {
	return &AwsS3Config{
		Region:          os.Getenv("TEST_S3_REGION"),
		Bucket:          os.Getenv("TEST_S3_BUCKET"),
		AccessKeyId:     os.Getenv("TEST_S3_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("TEST_S3_SECRET_ACCESS_KEY"),
	}
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

type FakeEELicense struct{}

func (f *FakeEELicense) IsValid() bool {
	return true
}
