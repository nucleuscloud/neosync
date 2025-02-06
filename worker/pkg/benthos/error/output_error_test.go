package neosync_benthos_error

import (
	"context"
	"testing"

	"github.com/redpanda-data/benthos/v4/public/service"
	"github.com/stretchr/testify/require"
)

func Test_ErrorOutputEmptyShutdown(t *testing.T) {
	conf := `
error_msg: "test error"
`
	spec := errorOutputSpec()
	env := service.NewEnvironment()

	config, err := spec.ParseYAML(conf, env)
	require.NoError(t, err)
	errorOutput, err := newErrorOutput(config, service.MockResources(), nil)
	require.NoError(t, err)
	require.NoError(t, errorOutput.Close(context.Background()))
}

func Test_ErrorOutputSendSignal(t *testing.T) {
	conf := `
error_msg: "${! meta(\"key\") }"
`
	spec := errorOutputSpec()
	env := service.NewEnvironment()

	config, err := spec.ParseYAML(conf, env)
	require.NoError(t, err)
	stopActivityChan := make(chan error, 1)
	errorOutput, err := newErrorOutput(config, service.MockResources(), stopActivityChan)
	require.NoError(t, err)
	msg := service.NewMessage([]byte("content"))
	msg.MetaSet("key", "duplicate key value violates unique constraint")

	batch := service.MessageBatch{msg}

	ctx := context.Background()
	err = errorOutput.WriteBatch(ctx, batch)
	require.NoError(t, err)
	out := <-stopActivityChan
	require.Error(t, out)
}

func Test_ErrorOutputMaxConnError(t *testing.T) {
	conf := `
error_msg: "${! meta(\"key\") }"
`
	spec := errorOutputSpec()
	env := service.NewEnvironment()

	config, err := spec.ParseYAML(conf, env)
	require.NoError(t, err)
	stopActivityChan := make(chan error, 1)
	errorOutput, err := newErrorOutput(config, service.MockResources(), stopActivityChan)
	require.NoError(t, err)
	msg := service.NewMessage([]byte("content"))
	msg.MetaSet("key", "too many clients already")

	batch := service.MessageBatch{msg}

	ctx := context.Background()
	err = errorOutput.WriteBatch(ctx, batch)
	require.Error(t, err)
}
