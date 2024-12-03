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
		Region: os.Getenv("TEST_S3_REGION"),
		Bucket: os.Getenv("TEST_S3_BUCKET"),
	}
}

func GetTestLogger(t testing.TB) *slog.Logger {
	f := slogt.Factory(func(w io.Writer) slog.Handler {
		return slog.NewTextHandler(w, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	})

	return slogt.New(t, f)
}

type FakeEELicense struct{}

func (f *FakeEELicense) IsValid() bool {
	return true
}
