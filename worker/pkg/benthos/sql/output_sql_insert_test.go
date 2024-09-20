package neosync_benthos_sql

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/warpstreamlabs/bento/public/service"
)

func Test_SqlInsertOutputEmptyShutdown(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
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

	insertOutput, err := newInsertOutput(insertConfig, service.MockResources(), nil, false, logger)
	require.NoError(t, err)
	require.NoError(t, insertOutput.Close(context.Background()))
}
