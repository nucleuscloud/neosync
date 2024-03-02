package neosync_benthos_sql

import (
	"context"
	"testing"

	"github.com/benthosdev/benthos/v4/public/service"
	"github.com/stretchr/testify/require"
)

func Test_SqlRawOutputEmptyShutdown(t *testing.T) {
	conf := `
driver: postgres
dsn: foo
query: "select * from public.users"
args_mapping: 'root = [this.id]'
`
	spec := sqlRawOutputSpec()
	env := service.NewEnvironment()

	insertConfig, err := spec.ParseYAML(conf, env)
	require.NoError(t, err)

	insertOutput, err := newOutput(insertConfig, service.MockResources(), nil)
	require.NoError(t, err)
	require.NoError(t, insertOutput.Close(context.Background()))
}
