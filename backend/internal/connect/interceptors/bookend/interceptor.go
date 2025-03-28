package bookend_logging_interceptor

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"connectrpc.com/connect"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/pkg/utils"
)

type Interceptor struct {
	config *config
}

func NewInterceptor(opts ...Option) connect.Interceptor {
	cfg := &config{
		level: slog.LevelInfo,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return &Interceptor{config: cfg}
}

type config struct {
	level slog.Level
}

type Option func(*config)

// Used to configure the standard level that the bookend messages are logged at.
// The error log is always printed at the Error level regardless of the level set.
func WithLogLevel(level slog.Level) Option {
	return func(o *config) {
		o.level = level
	}
}

func (i *Interceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
		logger = logger.With("procedure", request.Spec().Procedure)

		cliAttr := getCliAttr(request.Header())
		if cliAttr != nil {
			logger = logger.With(*cliAttr)
		}

		ctx = logger_interceptor.SetLoggerContext(ctx, logger)

		now := time.Now()
		logger.Log(
			ctx,
			i.config.level,
			"started call",
			"time_ms", fmt.Sprintf("%d", now.UnixMilli()),
			"stream_type", request.Spec().StreamType.String(),
			"http_method", request.HTTPMethod(),
			"peer_protocol", request.Peer().Protocol,
		)
		resp, err := next(ctx, request)
		endNow := time.Now()
		fields := []any{
			"time_ms", fmt.Sprintf("%d", endNow.UnixMilli()),
			"duration_ms", fmt.Sprintf("%d", endNow.Sub(now).Milliseconds()),
			"stream_type", request.Spec().StreamType.String(),
			"http_method", request.HTTPMethod(),
			"peer_protocol", request.Peer().Protocol,
		}

		if err != nil {
			fields = append(fields, "error", fmt.Sprintf("%v", err))
			connectErr, ok := err.(*connect.Error)
			if ok {
				fields = append(fields, "connect.code", connectErr.Code().String())
			} else {
				fields = append(fields, "connect.code", connect.CodeInternal.String())
			}

			logger.ErrorContext(ctx, err.Error(), fields...)

			logger.Log(
				ctx,
				i.config.level,
				"finished call",
				fields...,
			)
			return nil, err
		}
		fields = append(fields, "connect.code", "ok")
		logger.Log(
			ctx,
			i.config.level,
			"finished call",
			fields...,
		)
		return resp, nil
	}
}

func getCliAttr(header http.Header) *slog.Attr {
	cliVersion := header.Get(utils.CliVersionKey)
	cliPlatform := header.Get(utils.CliPlatformKey)
	cliCommit := header.Get(utils.CliCommitKey)

	attrs := []slog.Attr{}
	if cliVersion != "" {
		attrs = append(attrs, slog.String("version", cliVersion))
	}
	if cliPlatform != "" {
		attrs = append(attrs, slog.String("platform", cliPlatform))
	}
	if cliCommit != "" {
		attrs = append(attrs, slog.String("commit", cliCommit))
	}

	attrAny := make([]any, len(attrs))
	for i, attr := range attrs {
		attrAny[i] = attr
	}

	if len(attrAny) == 0 {
		return nil
	}
	cliGroup := slog.Group("cli", attrAny...)
	return &cliGroup
}

func (i *Interceptor) WrapStreamingClient(
	next connect.StreamingClientFunc,
) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		return next(ctx, spec)
	}
}

func (i *Interceptor) WrapStreamingHandler(
	next connect.StreamingHandlerFunc,
) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
		logger = logger.With("procedure", conn.Spec().Procedure)
		ctx = logger_interceptor.SetLoggerContext(ctx, logger)

		now := time.Now()
		logger.Info(
			"started call",
			"time_ms", fmt.Sprintf("%d", now.UnixMilli()),
			"stream_type", conn.Spec().StreamType.String(),
			"peer_protocol", conn.Peer().Protocol,
		)
		err := next(ctx, conn)
		endNow := time.Now()
		fields := []any{
			"time_ms", fmt.Sprintf("%d", endNow.UnixMilli()),
			"duration_ms", fmt.Sprintf("%d", endNow.Sub(now).Milliseconds()),
			"stream_type", conn.Spec().StreamType.String(),
			"peer_protocol", conn.Peer().Protocol,
		}
		if err != nil {
			fields = append(fields, "error", fmt.Sprintf("%v", err))
			connectErr, ok := err.(*connect.Error)
			if ok {
				fields = append(fields, "connect.code", connectErr.Code().String())
			}

			logger.Info(
				"finished call",
				fields...,
			)
			return err
		}
		fields = append(fields, "connect.code", "ok")
		logger.Info(
			"finished call",
			fields...,
		)
		return nil
	}
}
