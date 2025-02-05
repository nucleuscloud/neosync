package transformer_executor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_AnonValueApi_getPropertyPathValue(t *testing.T) {
	message, err := NewMessage(map[string]any{"a": "b"})
	require.NoError(t, err)
	api := newAnonValueApi()
	api.SetMessage(message)
	value, err := api.GetPropertyPathValue("a")
	require.NoError(t, err)
	require.Equal(t, "b", value)
}
