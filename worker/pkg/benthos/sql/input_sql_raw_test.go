package neosync_benthos_sql

import (
	"context"
	"fmt"
	"testing"

	"github.com/redpanda-data/benthos/v4/public/service"
	"github.com/stretchr/testify/require"
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

	selectInput, err := newInput(selectConfig, service.MockResources(), &fakeConnectionProvider{}, nil, nil)
	require.NoError(t, err)
	require.NoError(t, selectInput.Close(context.Background()))
}

type fakeConnectionProvider struct{}

func (f *fakeConnectionProvider) GetDriver(connectionId string) (string, error) {
	return "pgx", nil
}
func (f *fakeConnectionProvider) GetDb(ctx context.Context, connectionId string) (SqlDbtx, error) {
	return nil, fmt.Errorf("this is a test so if you need this, fix it or generate a mock")
}
