package testutil

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"
	"time"

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

func (f *FakeEELicense) ExpiresAt() time.Time {
	return time.Now().Add(time.Hour * 24 * 365)
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

func WithDockerFile(df *testcontainers.FromDockerfile) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.FromDockerfile = *df
		req.Image = ""
		return nil
	}
}

type TlsCertificatePathResponse struct {
	ServerCertPath string
	ServerKeyPath  string
	RootCertPath   string
	ClientCertPath string
	ClientKeyPath  string
}

const (
	// This is the relative path to where this mtls cert and Docker files live
	composeMtlsCertsRelativePath = "../../compose"
)

func GetTlsCertificatePaths() (*TlsCertificatePathResponse, error) {
	// when mounting files in testcontainers, they must be an absolute path
	basePath, err := resolveAbsolutePath(composeMtlsCertsRelativePath)
	if err != nil {
		return nil, err
	}
	return &TlsCertificatePathResponse{
		ServerCertPath: path.Join(basePath, "mtls/server/server.crt"),
		ServerKeyPath:  path.Join(basePath, "mtls/server/server.key"),
		RootCertPath:   path.Join(basePath, "mtls/ca/ca.crt"),
		ClientCertPath: path.Join(basePath, "mtls/client/client.crt"),
		ClientKeyPath:  path.Join(basePath, "mtls/client/client.key"),
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

// Mssql mtls must use a custom dockerfile as it's not possible to layer in certs into the running container
// due to sqlserver using a readonly filesystem
func GetMssqlTlsDockerfile() (*testcontainers.FromDockerfile, error) {
	basePath, err := resolveAbsolutePath(composeMtlsCertsRelativePath)
	if err != nil {
		return nil, err
	}
	return &testcontainers.FromDockerfile{
		Dockerfile: "Dockerfile.mssqlssl",
		Context:    basePath,
	}, nil
}

func GetClientTlsConfig(
	serverHost string,
) (*tls.Config, error) {
	certPaths, err := GetTlsCertificatePaths()
	if err != nil {
		return nil, err
	}
	cert, err := tls.LoadX509KeyPair(certPaths.ClientCertPath, certPaths.ClientKeyPath)
	if err != nil {
		return nil, err
	}

	rootCas := x509.NewCertPool()
	rootbits, err := os.ReadFile(certPaths.RootCertPath)
	if err != nil {
		return nil, err
	}
	ok := rootCas.AppendCertsFromPEM(rootbits)
	if !ok {
		return nil, errors.New("was unable to add test root cert to root ca pool")
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      rootCas,
		MinVersion:   tls.VersionTLS12,
		ServerName:   serverHost,
	}, nil
}
