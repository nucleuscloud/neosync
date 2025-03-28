package benthos_slogger

import (
	"context"
	"log/slog"

	"github.com/redpanda-data/benthos/v4/public/service"
)

var _ slog.Handler = (*benthosLogHandler)(nil)

type benthosLogHandler struct {
	logger *service.Logger
	attrs  []slog.Attr
	groups []string
}

func (h *benthosLogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	// We defer to the benthos logger and let it handle what leveling it wants to output
	return true
}

func (h *benthosLogHandler) Handle(
	ctx context.Context,
	r slog.Record,
) error { //nolint:gocritic // Needs to conform to the slog.Handler interface
	// Combine pre-defined attrs with record attrs
	allAttrs := make([]slog.Attr, 0, len(h.attrs)+r.NumAttrs())
	allAttrs = append(allAttrs, h.attrs...)

	r.Attrs(func(attr slog.Attr) bool {
		if !attr.Equal(slog.Attr{}) {
			// Handle groups
			if len(h.groups) > 0 {
				last := h.groups[len(h.groups)-1]
				if last != "" {
					attr.Key = last + "." + attr.Key
				}
			}
			allAttrs = append(allAttrs, attr)
		}
		return true
	})

	// Convert to key-value pairs for temporal logger
	keyvals := make([]any, 0, len(allAttrs)*2)
	for _, attr := range allAttrs {
		keyvals = append(keyvals, attr.Key, attr.Value.Any())
	}

	switch r.Level {
	case slog.LevelDebug:
		h.logger.With(keyvals...).Debug(r.Message)
	case slog.LevelInfo:
		h.logger.With(keyvals...).Info(r.Message)
	case slog.LevelWarn:
		h.logger.With(keyvals...).Warn(r.Message)
	case slog.LevelError:
		h.logger.With(keyvals...).Error(r.Message)
	}
	return nil
}

func (h *benthosLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := []slog.Attr{}
	newAttrs = append(newAttrs, h.attrs...)
	newAttrs = append(newAttrs, attrs...)
	return &benthosLogHandler{
		logger: h.logger,
		attrs:  newAttrs,
		groups: h.groups,
	}
}

func (h *benthosLogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	newGroups := []string{}
	newGroups = append(newGroups, h.groups...)
	newGroups = append(newGroups, name)
	return &benthosLogHandler{
		logger: h.logger,
		attrs:  h.attrs,
		groups: newGroups,
	}
}

func newBenthosLogHandler(logger *service.Logger) *benthosLogHandler {
	return &benthosLogHandler{logger: logger}
}

// Returns a benthos logger wrapped as a slog.Logger to ease plugging in to the rest of the system
func NewSlogger(logger *service.Logger) *slog.Logger {
	return slog.New(newBenthosLogHandler(logger))
}
