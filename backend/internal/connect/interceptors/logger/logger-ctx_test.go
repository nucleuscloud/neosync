package logger_interceptor

import (
	"context"
	"testing"

	"github.com/nucleuscloud/neosync/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func Test_GetLoggerFromContextOrDefault(t *testing.T) {
	assert.NotNil(t, GetLoggerFromContextOrDefault(context.Background()))
}

func Test_GetLoggerFromContextOrDefault_NonDefault(t *testing.T) {
	logger := testutil.GetTestLogger(t)
	ctx := SetLoggerContext(context.Background(), logger)
	ctxlogger := GetLoggerFromContextOrDefault(ctx)
	assert.Equal(t, logger, ctxlogger)
}
