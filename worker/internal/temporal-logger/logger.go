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
	// We defer to the temporal logger and let it handle what leveling it wants to output
	return true
}

func (h *temporalLogHandler) Handle(
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
		h.logger.Debug(r.Message, keyvals...)
	case slog.LevelInfo:
		h.logger.Info(r.Message, keyvals...)
	case slog.LevelWarn:
		h.logger.Warn(r.Message, keyvals...)
	case slog.LevelError:
		h.logger.Error(r.Message, keyvals...)
	}
	return nil
}

func (h *temporalLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := []slog.Attr{}
	newAttrs = append(newAttrs, h.attrs...)
	newAttrs = append(newAttrs, attrs...)
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
	newGroups := []string{}
	newGroups = append(newGroups, h.groups...)
	newGroups = append(newGroups, name)
	return &temporalLogHandler{
		logger: h.logger,
		attrs:  h.attrs,
		groups: newGroups,
	}
}

func newTemporalLogHandler(tlogger log.Logger) *temporalLogHandler {
	return &temporalLogHandler{logger: tlogger}
}

// Returns a temporal logger wrapped as a slog.Logger to ease plugging in to the rest of the system
func NewSlogger(tlogger log.Logger) *slog.Logger {
	handler := newTemporalLogHandler(tlogger)
	return slog.New(handler)
}
