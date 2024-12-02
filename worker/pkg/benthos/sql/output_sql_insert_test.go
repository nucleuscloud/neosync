package neosync_benthos_sql

import (
	"context"
	"testing"

	"github.com/nucleuscloud/neosync/internal/testutil"
	"github.com/stretchr/testify/require"
	"github.com/warpstreamlabs/bento/public/service"
)

func Test_SqlInsertOutputEmptyShutdown(t *testing.T) {
	conf := `
connection_id: 123
schema: bar
table: baz
args_mapping: 'root = [this.id]'
`
	spec := sqlInsertOutputSpec()
	env := service.NewEnvironment()

	insertConfig, err := spec.ParseYAML(conf, env)
	require.NoError(t, err)

	insertOutput, err := newInsertOutput(insertConfig, service.MockResources(), &fakeConnectionProvider{}, false, testutil.GetTestLogger(t))
	require.NoError(t, err)
	require.NoError(t, insertOutput.Close(context.Background()))
}
