package neosync_benthos_sql

import (
	"context"
	"testing"

	"github.com/benthosdev/benthos/v4/public/service"
	"github.com/stretchr/testify/require"
)

func Test_SqlInsertOutputEmptyShutdown(t *testing.T) {
	conf := `
driver: postgres
dsn: foo
schema: bar
table: baz
args_mapping: 'root = [this.id]'
`
	spec := sqlInsertOutputSpec()
	env := service.NewEnvironment()

	insertConfig, err := spec.ParseYAML(conf, env)
	require.NoError(t, err)

	insertOutput, err := newInsertOutput(insertConfig, service.MockResources(), nil)
	require.NoError(t, err)
	require.NoError(t, insertOutput.Close(context.Background()))
}
