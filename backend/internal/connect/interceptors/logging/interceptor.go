package logging_interceptor

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"connectrpc.com/connect"
)

type Interceptor struct {
	logger *slog.Logger
}

func NewInterceptor(logger *slog.Logger) connect.Interceptor {
	return &Interceptor{logger: logger}
}

func (i *Interceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		now := time.Now()
		i.logger.Info(
			"started call",
			"time_ms", fmt.Sprintf("%d", now.UnixMilli()),
			"stream_type", request.Spec().StreamType.String(),
			"procedure", request.Spec().Procedure,
			"http_method", request.HTTPMethod(),
			"peer_protocol", request.Peer().Protocol,
		)
		resp, err := next(ctx, request)
		endNow := time.Now()
		fields := []any{
			"time_ms", fmt.Sprintf("%d", endNow.UnixMilli()),
			"duration_ms", fmt.Sprintf("%d", endNow.Sub(now).Milliseconds()),
			"stream_type", request.Spec().StreamType.String(),
			"procedure", request.Spec().Procedure,
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

			i.logger.Error(err.Error())

			i.logger.Info(
				"finished call",
				fields...,
			)
			return nil, err
		}
		fields = append(fields, "connect.code", "ok")
		i.logger.Info(
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
		now := time.Now()
		i.logger.Info(
			"started call",
			"time_ms", fmt.Sprintf("%d", now.UnixMilli()),
			"stream_type", conn.Spec().StreamType.String(),
			"procedure", conn.Spec().Procedure,
			"peer_protocol", conn.Peer().Protocol,
		)
		err := next(ctx, conn)
		endNow := time.Now()
		fields := []any{
			"time_ms", fmt.Sprintf("%d", endNow.UnixMilli()),
			"duration_ms", fmt.Sprintf("%d", endNow.Sub(now).Milliseconds()),
			"stream_type", conn.Spec().StreamType.String(),
			"procedure", conn.Spec().Procedure,
			"peer_protocol", conn.Peer().Protocol,
		}
		if err != nil {
			fields = append(fields, "error", fmt.Sprintf("%v", err))
			connectErr, ok := err.(*connect.Error)
			if ok {
				fields = append(fields, "connect.code", connectErr.Code().String())
			}

			i.logger.Info(
				"finished call",
				fields...,
			)
			return err
		}
		fields = append(fields, "connect.code", "ok")
		i.logger.Info(
			"finished call",
			fields...,
		)
		return nil
	}
}
