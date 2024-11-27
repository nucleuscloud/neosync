package neosync_benthos_sql

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/warpstreamlabs/bento/public/service"
)

func Test_SqlUpdateOutputEmptyShutdown(t *testing.T) {
	conf := `
connection_id: 123
schema: bar
table: baz
args_mapping: 'root = [this.id]'
`
	spec := sqlUpdateOutputSpec()
	env := service.NewEnvironment()

	updateConfig, err := spec.ParseYAML(conf, env)
	require.NoError(t, err)

	updateOutput, err := newUpdateOutput(updateConfig, service.MockResources(), nil)
	require.NoError(t, err)
	require.NoError(t, updateOutput.Close(context.Background()))
}
