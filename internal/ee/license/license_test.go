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
	validExpiredTestLicense = "eyJjb250ZW50cyI6ImV3b2dJQ0FnSW5abGNuTnBiMjRpT2lBaWRqRWlMQW9nSUNBZ0ltbGtJam9nSWpFeU15SXNDaUFnSUNBaVpYaHdhWEo1WDJSaGRHVWlPaUFpTWpBeU15MHhNaTB6TVZReE1qb3dNRG93TUZvaUNuMEsiLCJzaWduYXR1cmUiOiJuRVp3TFJqaHhtMXFsMHpNalBsRHkvV0gwN3RBeUsvUDBOejg5bjNkZ3ZXY3ZqWG5oVGFxZWcvazNaRTJKNlk0d3NyVzh2M2x6M2UxdmY5d3poZGNDQT09In0K"
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
		require.Equal(t, time.Date(2023, 12, 31, 12, 0, 0, 0, time.UTC), contents.ExpiryDate)
	})
}

func Test_NewFromEnv(t *testing.T) {
	viper.Set(eeLicenseEvKey, validExpiredTestLicense)
	eelicense, err := NewFromEnv()
	require.NoError(t, err)
	require.NotNil(t, eelicense)

	require.False(t, eelicense.IsValid())
}
