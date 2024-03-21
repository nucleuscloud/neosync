package neosync_benthos_error

import (
	"context"
	"testing"

	"github.com/benthosdev/benthos/v4/public/service"
	"github.com/stretchr/testify/require"
)

func Test_ErrorProcessorEmptyShutdown(t *testing.T) {
	errorProcessor := newErrorProcessor(service.MockResources().Logger(), nil)
	require.NoError(t, errorProcessor.Close(context.Background()))
}

func Test_ErrorProcessorSendSignal(t *testing.T) {
	stopWorkflowChan := make(chan bool, 1)
	errorProcessor := newErrorProcessor(service.MockResources().Logger(), stopWorkflowChan)
	ctx := context.Background()
	_, err := errorProcessor.ProcessBatch(ctx, service.MessageBatch{})
	out := <-stopWorkflowChan
	require.True(t, out)
	require.Error(t, err)
}
