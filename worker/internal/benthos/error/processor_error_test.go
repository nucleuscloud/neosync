package neosync_benthos_error

import (
	"context"
	"fmt"
	"testing"

	"github.com/benthosdev/benthos/v4/public/service"
	"github.com/stretchr/testify/require"
)

func Test_ErrorProcessorEmptyShutdown(t *testing.T) {
	errorProcessor := newErrorProcessor(service.MockResources().Logger(), nil)
	require.NoError(t, errorProcessor.Close(context.Background()))
}

func Test_ErrorProcessorSendSignal(t *testing.T) {
	stopActivityChan := make(chan error, 1)
	errorProcessor := newErrorProcessor(service.MockResources().Logger(), stopActivityChan)
	ctx := context.Background()
	_, err := errorProcessor.ProcessBatch(ctx, service.MessageBatch{})
	require.NoError(t, err)
	out := <-stopActivityChan
	require.Equal(t, out, fmt.Errorf("Processor Error"))
}
