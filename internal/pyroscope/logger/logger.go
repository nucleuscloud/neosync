package pyroscope_logger

import (
	"fmt"
	"log/slog"

	"github.com/grafana/pyroscope-go"
)

type PyroscopeLogger struct {
	logger *slog.Logger
}

var _ pyroscope.Logger = (*PyroscopeLogger)(nil)

func (p *PyroscopeLogger) Debugf(format string, args ...any) {
	p.logger.Debug(fmt.Sprintf(format, args...))
}

func (p *PyroscopeLogger) Infof(format string, args ...any) {
	p.logger.Info(fmt.Sprintf(format, args...))
}

func (p *PyroscopeLogger) Errorf(format string, args ...any) {
	p.logger.Error(fmt.Sprintf(format, args...))
}

func New(logger *slog.Logger) *PyroscopeLogger {
	return &PyroscopeLogger{logger: logger}
}

type noopLogger struct{}

var _ pyroscope.Logger = (*noopLogger)(nil)

func (n *noopLogger) Debugf(_ string, _ ...any) {}
func (n *noopLogger) Infof(_ string, _ ...any)  {}
func (n *noopLogger) Errorf(_ string, _ ...any) {}

func NewNoop() *noopLogger {
	return &noopLogger{}
}
