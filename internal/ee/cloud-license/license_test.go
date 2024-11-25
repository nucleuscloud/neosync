package cloudlicense

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
	// ./scripts/gen-cust-license.sh ./neosync_cloud_ca.key license.json | pbcopy
	validExpiredTestLicense = "eyJsaWNlbnNlIjoiZXdvZ0lDQWdJblpsY25OcGIyNGlPaUFpZGpFaUxBb2dJQ0FnSW1sa0lqb2dJbVk0TW1aaVlXWmtMVFppTnpVdE5HSXpaUzFoWmpRekxUZGhaRFF3TldNNFpEUTRZaUlzQ2lBZ0lDQWlhWE56ZFdWa1gzUnZJam9nSWtGamJXVWdRMjh1SWl3S0lDQWdJQ0pwYzNOMVpXUmZZWFFpT2lBaU1qQXlNaTB4TWkwek1WUXhNam93TURvd01Gb2lMQW9nSUNBZ0ltVjRjR2x5WlhOZllYUWlPaUFpTWpBeU15MHhNaTB6TVZReE1qb3dNRG93TUZvaUNuMEsiLCJzaWduYXR1cmUiOiJMOWxTT3dkL2VjMmlpZVlYYUFSRENlUzhtaE5INS85c1M0VHQvNkJVMHJmQXMraTRLYVJRV1p5eG9Id203eC8vb2VReXd4cmN1VGpQUXFvemFHbHJEdz09In0K"
)

func Test_getLicense(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		publicKey, err := parsePublicKey(publicKeyPEM)
		require.NoError(t, err)
		contents, err := getLicense(validExpiredTestLicense, publicKey)
		require.NoError(t, err)
		require.NotEmpty(t, contents)

		require.Equal(t, "f82fbafd-6b75-4b3e-af43-7ad405c8d48b", contents.Id)
		require.Equal(t, "v1", contents.Version)
		require.Equal(t, time.Date(2023, 12, 31, 12, 0, 0, 0, time.UTC), contents.ExpiresAt)
		require.Equal(t, time.Date(2022, 12, 31, 12, 0, 0, 0, time.UTC), contents.IssuedAt)
		require.Equal(t, "Acme Co.", contents.IssuedTo)
		require.False(t, contents.IsValid())
	})
}

func Test_NewFromEnv(t *testing.T) {
	t.Run("present", func(t *testing.T) {
		viper.Set(cloudLicenseEvKey, validExpiredTestLicense)
		eelicense, err := NewFromEnv()
		require.NoError(t, err)
		require.NotNil(t, eelicense)

		require.False(t, eelicense.IsValid())
	})
	t.Run("empty", func(t *testing.T) {
		viper.Set(cloudLicenseEvKey, "")

		viper.Set(cloudLicenseEvKey, validExpiredTestLicense)
		eelicense, err := NewFromEnv()
		require.NoError(t, err)
		require.NotNil(t, eelicense)

		require.False(t, eelicense.IsValid())
	})
}
