package testutil

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/neilotoole/slogt"
	"github.com/testcontainers/testcontainers-go"
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

// not safe for concurrent use
func GetTestLogger(t testing.TB) *slog.Logger {
	f := slogt.Factory(func(w io.Writer) slog.Handler {
		return slog.NewTextHandler(w, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	})

	return slogt.New(t, f)
}

type FakeEELicense struct {
	isValid bool
}

type Option func(*FakeEELicense)

func WithIsValid() Option {
	return func(f *FakeEELicense) {
		f.isValid = true
	}
}

func NewFakeEELicense(opts ...Option) *FakeEELicense {
	f := &FakeEELicense{}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

func (f *FakeEELicense) IsValid() bool {
	return f.isValid
}

func GetConcurrentTestLogger(t testing.TB) *slog.Logger {
	if testing.Verbose() {
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	}
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
}

func WithFiles(files []testcontainers.ContainerFile) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Files = files
		return nil
	}
}

func WithCmd(cmd []string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Cmd = cmd
		return nil
	}
}

func WithDockerFile(df testcontainers.FromDockerfile) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.FromDockerfile = df
		return nil
	}
}

type TlsServerCertificateResponse struct {
	ServerCrtPath string
	ServerKeyPath string
	RootCrtPath   string
}

const (
	tlsCertsRelativePath = "../../compose/mtls"
)

func GetTlsServerCertificatePaths() (*TlsServerCertificateResponse, error) {
	// when mounting files in testcontainers, they must be an absolute path
	basePath, err := resolveAbsolutePath(tlsCertsRelativePath)
	if err != nil {
		return nil, err
	}
	return &TlsServerCertificateResponse{
		ServerCrtPath: path.Join(basePath, "server/server.crt"),
		ServerKeyPath: path.Join(basePath, "server/server.key"),
		RootCrtPath:   path.Join(basePath, "ca/ca.crt"),
	}, nil
}

func resolveAbsolutePath(relpath string) (string, error) {
	// Get current file path
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("failed to get current file path")
	}

	certsPath := filepath.Join(filepath.Dir(filename), relpath)
	absPath, err := filepath.Abs(certsPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}
	return absPath, nil
}
