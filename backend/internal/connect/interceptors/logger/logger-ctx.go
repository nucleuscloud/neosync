package logger_interceptor

import (
	"context"
	"log/slog"
)

type loggerContextKey struct{}
type loggerContextData struct {
	logger *slog.Logger
}

func (l *loggerContextData) GetLogger() *slog.Logger {
	return l.logger
}

func GetLoggerFromContextOrDefault(ctx context.Context) *slog.Logger {
	data, ok := ctx.Value(loggerContextKey{}).(*loggerContextData)
	if !ok {
		return slog.Default()
	}
	return data.GetLogger()
}

func setLoggerContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey{}, &loggerContextData{logger: logger})
}
