package javascript_vm

import (
	"fmt"
	"log/slog"

	"github.com/dop251/goja_nodejs/console"
)

var _ console.Printer = &consoleLogger{}

// adds a standard prefix to the message to make it easier to identify logs that originate from the JS VM.
const stdPrefix = "[js]: "

func newConsoleLogger(prefix string, logger *slog.Logger) *consoleLogger {
	return &consoleLogger{prefix: prefix, logger: logger}
}

type consoleLogger struct {
	prefix string
	logger *slog.Logger
}

func (l *consoleLogger) Log(message string) {
	l.logger.Info(fmt.Sprintf("%s%s", l.prefix, message))
}

func (l *consoleLogger) Warn(message string) {
	l.logger.Warn(fmt.Sprintf("%s%s", l.prefix, message))
}

func (l *consoleLogger) Error(message string) {
	l.logger.Error(fmt.Sprintf("%s%s", l.prefix, message))
}
