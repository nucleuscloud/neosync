package logging_interceptor

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
)

type Interceptor struct {
}

func NewInterceptor() connect.Interceptor {
	return &Interceptor{}
}

func (i *Interceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
		logger = logger.With("procedure", request.Spec().Procedure)
		ctx = logger_interceptor.SetLoggerContext(ctx, logger)

		now := time.Now()
		logger.Info(
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

			logger.Error(err.Error(), fields...)

			logger.Info(
				"finished call",
				fields...,
			)
			return nil, err
		}
		fields = append(fields, "connect.code", "ok")
		logger.Info(
			"finished call",
			fields...,
		)
		return resp, nil
	}
}

func (i *Interceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		return next(ctx, spec)
	}
}

func (i *Interceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
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
