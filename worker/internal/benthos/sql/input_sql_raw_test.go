package sql

import (
	"context"
	"testing"

	"github.com/benthosdev/benthos/v4/public/service"
	"github.com/stretchr/testify/require"
)

func Test_SqlRawInputEmptyShutdown(t *testing.T) {
	conf := `
query: "select * from public.users"
args_mapping: 'root = [this.id]'
`
	spec := sqlRawInputSpec()
	env := service.NewEnvironment()

	selectConfig, err := spec.ParseYAML(conf, env)
	require.NoError(t, err)

	selectInput, err := newInput(selectConfig, service.MockResources())
	require.NoError(t, err)
	require.NoError(t, selectInput.Close(context.Background()))
}
