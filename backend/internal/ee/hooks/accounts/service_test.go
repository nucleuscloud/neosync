package accounthooks

import (
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/stretchr/testify/require"
)

func Test_hasSlackChannelIdChanged(t *testing.T) {
	t.Run("nil hooks", func(t *testing.T) {
		result := hasSlackChannelIdChanged(nil, nil)
		require.False(t, result, "hasSlackChannelIdChanged() with nil hooks should return false")
	})

	t.Run("old hook nil", func(t *testing.T) {
		newHook := &mgmtv1alpha1.AccountHook{}
		result := hasSlackChannelIdChanged(nil, newHook)
		require.False(t, result, "hasSlackChannelIdChanged() with old hook nil should return false")
	})

	t.Run("new hook nil", func(t *testing.T) {
		oldHook := &mgmtv1alpha1.AccountHook{}
		result := hasSlackChannelIdChanged(oldHook, nil)
		require.False(t, result, "hasSlackChannelIdChanged() with new hook nil should return false")
	})

	t.Run("same channel id", func(t *testing.T) {
		oldHook := &mgmtv1alpha1.AccountHook{
			Config: &mgmtv1alpha1.AccountHookConfig{
				Config: &mgmtv1alpha1.AccountHookConfig_Slack{
					Slack: &mgmtv1alpha1.AccountHookConfig_SlackHook{
						ChannelId: "C123456",
					},
				},
			},
		}
		newHook := &mgmtv1alpha1.AccountHook{
			Config: &mgmtv1alpha1.AccountHookConfig{
				Config: &mgmtv1alpha1.AccountHookConfig_Slack{
					Slack: &mgmtv1alpha1.AccountHookConfig_SlackHook{
						ChannelId: "C123456",
					},
				},
			},
		}
		result := hasSlackChannelIdChanged(oldHook, newHook)
		require.False(t, result, "hasSlackChannelIdChanged() with same channel ID should return false")
	})

	t.Run("different channel id", func(t *testing.T) {
		oldHook := &mgmtv1alpha1.AccountHook{
			Config: &mgmtv1alpha1.AccountHookConfig{
				Config: &mgmtv1alpha1.AccountHookConfig_Slack{
					Slack: &mgmtv1alpha1.AccountHookConfig_SlackHook{
						ChannelId: "C123456",
					},
				},
			},
		}
		newHook := &mgmtv1alpha1.AccountHook{
			Config: &mgmtv1alpha1.AccountHookConfig{
				Config: &mgmtv1alpha1.AccountHookConfig_Slack{
					Slack: &mgmtv1alpha1.AccountHookConfig_SlackHook{
						ChannelId: "C789012",
					},
				},
			},
		}
		result := hasSlackChannelIdChanged(oldHook, newHook)
		require.True(t, result, "hasSlackChannelIdChanged() with different channel ID should return true")
	})

	t.Run("old hook missing slack config", func(t *testing.T) {
		oldHook := &mgmtv1alpha1.AccountHook{
			Config: &mgmtv1alpha1.AccountHookConfig{},
		}
		newHook := &mgmtv1alpha1.AccountHook{
			Config: &mgmtv1alpha1.AccountHookConfig{
				Config: &mgmtv1alpha1.AccountHookConfig_Slack{
					Slack: &mgmtv1alpha1.AccountHookConfig_SlackHook{
						ChannelId: "C123456",
					},
				},
			},
		}
		result := hasSlackChannelIdChanged(oldHook, newHook)
		require.True(t, result, "hasSlackChannelIdChanged() with old hook missing slack config should return true")
	})

	t.Run("new hook missing slack config", func(t *testing.T) {
		oldHook := &mgmtv1alpha1.AccountHook{
			Config: &mgmtv1alpha1.AccountHookConfig{
				Config: &mgmtv1alpha1.AccountHookConfig_Slack{
					Slack: &mgmtv1alpha1.AccountHookConfig_SlackHook{
						ChannelId: "C123456",
					},
				},
			},
		}
		newHook := &mgmtv1alpha1.AccountHook{
			Config: &mgmtv1alpha1.AccountHookConfig{},
		}
		result := hasSlackChannelIdChanged(oldHook, newHook)
		require.True(t, result, "hasSlackChannelIdChanged() with new hook missing slack config should return true")
	})
}
