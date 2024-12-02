package temporallogger

import (
	"context"
	"log/slog"

	"go.temporal.io/sdk/log"
)

var _ slog.Handler = (*temporalLogHandler)(nil)

type temporalLogHandler struct {
	logger log.Logger
	attrs  []slog.Attr
	groups []string
}

func (h *temporalLogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	// Temporal logger doesn't provide level filtering, so we accept all levels
	return true
}

func (h *temporalLogHandler) Handle(ctx context.Context, r slog.Record) error { //nolint:gocritic
	attrs := make([]any, 0, (r.NumAttrs()+len(h.attrs))*2)

	// Add handler's attrs first
	for _, attr := range h.attrs {
		attrs = append(attrs, attr.Key, attr.Value.String())
	}

	// Add record's attrs
	r.Attrs(func(a slog.Attr) bool {
		if !a.Equal(slog.Attr{}) {
			key := a.Key
			// Apply groups to key
			for _, group := range h.groups {
				if group != "" {
					key = group + "." + key
				}
			}
			attrs = append(attrs, key, a.Value.String())
		}
		return true
	})

	switch r.Level {
	case slog.LevelDebug:
		h.logger.Debug(r.Message, attrs...)
	case slog.LevelInfo:
		h.logger.Info(r.Message, attrs...)
	case slog.LevelWarn:
		h.logger.Warn(r.Message, attrs...)
	case slog.LevelError:
		h.logger.Error(r.Message, attrs...)
	}
	return nil
}

func (h *temporalLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)

	return &temporalLogHandler{
		logger: h.logger,
		attrs:  newAttrs,
		groups: h.groups,
	}
}

func (h *temporalLogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = name

	return &temporalLogHandler{
		logger: h.logger,
		attrs:  h.attrs,
		groups: newGroups,
	}
}

func NewSlogger(tlogger log.Logger) *slog.Logger {
	handler := &temporalLogHandler{logger: tlogger}
	return slog.New(handler)
}
