package neosync_benthos_sql

import (
	"context"
	"testing"

	"github.com/benthosdev/benthos/v4/public/service"
	"github.com/stretchr/testify/require"
)

func Test_SqlUpdateOutputEmptyShutdown(t *testing.T) {
	conf := `
driver: postgres
dsn: foo
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
