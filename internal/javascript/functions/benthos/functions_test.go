package benthos_functions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	functions := Get()
	require.NotEmpty(t, functions)
}
