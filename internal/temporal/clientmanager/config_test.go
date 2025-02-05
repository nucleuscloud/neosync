package clientmanager

import (
	"crypto/tls"
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/stretchr/testify/require"
)

func Test_Config_Override(t *testing.T) {
	t.Run("none", func(t *testing.T) {
		input := &TemporalConfig{
			Url:              "foo",
			Namespace:        "bar",
			SyncJobQueueName: "baz",
			TLSConfig:        &tls.Config{ServerName: "foo"},
		}
		output := input.Override(nil)
		require.True(t, input.Equals(output))
	})
	t.Run("all", func(t *testing.T) {
		input := &TemporalConfig{
			Url:              "foo",
			Namespace:        "bar",
			SyncJobQueueName: "baz",
			TLSConfig:        &tls.Config{ServerName: "foo"},
		}
		output := input.Override(&TemporalConfig{
			Url:              "foo1",
			Namespace:        "bar1",
			SyncJobQueueName: "baz1",
			TLSConfig:        &tls.Config{ServerName: "foo1"},
		})
		require.False(t, input.Equals(output))
		require.Equal(t, "foo1", output.Url)
		require.Equal(t, "bar1", output.Namespace)
		require.Equal(t, "baz1", output.SyncJobQueueName)
		require.Equal(t, "foo1", output.TLSConfig.ServerName)
	})
}
func Test_Config_ToDto(t *testing.T) {
	input := &TemporalConfig{
		Url:              "foo",
		Namespace:        "bar",
		SyncJobQueueName: "baz",
		TLSConfig:        &tls.Config{ServerName: "foo"},
	}
	output := input.ToDto()
	require.Equal(t, &mgmtv1alpha1.AccountTemporalConfig{
		Url:              "foo",
		Namespace:        "bar",
		SyncJobQueueName: "baz",
	}, output)
}
