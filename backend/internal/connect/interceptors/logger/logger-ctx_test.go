package logger_interceptor

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetLoggerFromContextOrDefault(t *testing.T) {
	assert.NotNil(t, GetLoggerFromContextOrDefault(context.Background()))
}

func Test_GetLoggerFromContextOrDefault_NonDefault(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := setLoggerContext(context.Background(), logger)
	ctxlogger := GetLoggerFromContextOrDefault(ctx)
	assert.Equal(t, logger, ctxlogger)
}
