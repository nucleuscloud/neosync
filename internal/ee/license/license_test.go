package license

import (
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func Test_parsePublicKey(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		actual, err := parsePublicKey("")
		require.Error(t, err)
		require.Nil(t, actual)
	})
	t.Run("invalid format", func(t *testing.T) {
		actual, err := parsePublicKey("blah")
		require.Error(t, err)
		require.Nil(t, actual)
	})
	t.Run("valid", func(t *testing.T) {
		actual, err := parsePublicKey(publicKeyPEM)
		require.NoError(t, err)
		require.NotNil(t, actual)
	})
}

const (
	// generated using the gen-cust-license shell script with the neosync ee private key
	// ./scripts/gen-cust-license.sh ./neosync_ee_ca.key license.json | pbcopy
	validExpiredTestLicense = "eyJsaWNlbnNlIjoiZXdvZ0lDQWdJblpsY25OcGIyNGlPaUFpZGpFaUxBb2dJQ0FnSW1sa0lqb2dJakV5TXlJc0NpQWdJQ0FpWlhod2FYSmxjMTloZENJNklDSXlNREl6TFRFeUxUTXhWREV5T2pBd09qQXdXaUlzQ2lBZ0lDQWlhWE56ZFdWa1gyRjBJam9nSWpJd01qSXRNVEl0TXpGVU1USTZNREE2TURCYUlpd0tJQ0FnSUNKcGMzTjFaV1JmZEc4aU9pQWlRV050WlNCRGJ5NGlMQW9nSUNBZ0ltTjFjM1J2YldWeVgybGtJam9nSWpRMU5pSUtmUW89Iiwic2lnbmF0dXJlIjoiZnJDNHlFUWdLSUFmT05SeHlRYVpXRHFqenpmYmpxS1ErazdKeCt2M2VNQ2dSRzhOQ05ITzYwVmE0M0VQZEVIdm1NVzVlMzk1cjBaK3lvUWhGM3RDRHc9PSJ9Cg=="
)

func Test_getLicense(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		publicKey, err := parsePublicKey(publicKeyPEM)
		require.NoError(t, err)
		contents, err := getLicense(validExpiredTestLicense, publicKey)
		require.NoError(t, err)
		require.NotEmpty(t, contents)

		require.Equal(t, "123", contents.Id)
		require.Equal(t, "v1", contents.Version)
		require.Equal(t, time.Date(2023, 12, 31, 12, 0, 0, 0, time.UTC), contents.ExpiresAt)
		require.Equal(t, time.Date(2022, 12, 31, 12, 0, 0, 0, time.UTC), contents.IssuedAt)
		require.Equal(t, "456", contents.CustomerId)
		require.Equal(t, "Acme Co.", contents.IssuedTo)
		require.False(t, contents.IsValid())
	})
}

func Test_NewFromEnv(t *testing.T) {
	t.Run("present", func(t *testing.T) {
		viper.Set(eeLicenseEvKey, validExpiredTestLicense)
		eelicense, err := NewFromEnv()
		require.NoError(t, err)
		require.NotNil(t, eelicense)

		require.False(t, eelicense.IsValid())
	})
	t.Run("empty", func(t *testing.T) {
		viper.Set(eeLicenseEvKey, "")

		viper.Set(eeLicenseEvKey, validExpiredTestLicense)
		eelicense, err := NewFromEnv()
		require.NoError(t, err)
		require.NotNil(t, eelicense)

		require.False(t, eelicense.IsValid())
	})
}
