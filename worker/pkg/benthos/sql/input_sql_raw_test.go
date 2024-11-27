package neosync_benthos_sql

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/warpstreamlabs/bento/public/service"
)

func Test_SqlRawInputEmptyShutdown(t *testing.T) {
	conf := `
connection_id: 123
query: "select * from public.users"
args_mapping: 'root = [this.id]'
`
	spec := sqlRawInputSpec()
	env := service.NewEnvironment()

	selectConfig, err := spec.ParseYAML(conf, env)
	require.NoError(t, err)

	selectInput, err := newInput(selectConfig, service.MockResources(), &fakeConnectionProvider{}, nil)
	require.NoError(t, err)
	require.NoError(t, selectInput.Close(context.Background()))
}

type fakeConnectionProvider struct{}

func (f *fakeConnectionProvider) GetDriver(connectionId string) (string, error) {
	return "postgres", nil
}
func (f *fakeConnectionProvider) GetDb(ctx context.Context, connectionId string) (SqlDbtx, error) {
	return nil, fmt.Errorf("this is a test so if you need this, fix it or generate a mock")
}
