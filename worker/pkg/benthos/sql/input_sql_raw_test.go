package neosync_benthos_sql

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/warpstreamlabs/bento/public/service"
)

func Test_SqlRawInputEmptyShutdown(t *testing.T) {
	conf := `
driver: postgres
dsn: foo
query: "select * from public.users"
args_mapping: 'root = [this.id]'
`
	spec := sqlRawInputSpec()
	env := service.NewEnvironment()

	selectConfig, err := spec.ParseYAML(conf, env)
	require.NoError(t, err)

	selectInput, err := newInput(selectConfig, service.MockResources(), nil, nil)
	require.NoError(t, err)
	require.NoError(t, selectInput.Close(context.Background()))
}
