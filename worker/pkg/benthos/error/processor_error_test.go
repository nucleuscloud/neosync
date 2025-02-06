package neosync_benthos_error

import (
	"context"
	"fmt"
	"testing"

	"github.com/redpanda-data/benthos/v4/public/service"
	"github.com/stretchr/testify/require"
)

func Test_ErrorProcessorEmptyShutdown(t *testing.T) {
	conf := `
error_msg: "test error"
`
	spec := errorOutputSpec()
	env := service.NewEnvironment()

	config, err := spec.ParseYAML(conf, env)
	require.NoError(t, err)
	errorProcessor, err := newErrorProcessor(config, service.MockResources(), nil)
	require.NoError(t, err)
	require.NoError(t, errorProcessor.Close(context.Background()))
}

func Test_ErrorProcessorSendSignal(t *testing.T) {
	conf := `
error_msg: "${! meta(\"key\") }"
`
	spec := errorOutputSpec()
	env := service.NewEnvironment()

	config, err := spec.ParseYAML(conf, env)
	require.NoError(t, err)
	stopActivityChan := make(chan error, 1)
	errorProcessor, err := newErrorProcessor(config, service.MockResources(), stopActivityChan)
	require.NoError(t, err)
	msg := service.NewMessage([]byte("content"))
	msg.MetaSet("key", "Processor Error")

	batch := service.MessageBatch{msg}

	ctx := context.Background()
	_, err = errorProcessor.ProcessBatch(ctx, batch)
	require.NoError(t, err)
	out := <-stopActivityChan
	require.Equal(t, out, fmt.Errorf("Processor Error"))
}
