package temporallogger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockTemporalLogger implements temporal's log.Logger interface for testing
type MockTemporalLogger struct {
	messages []LogMessage
}

type LogMessage struct {
	Level   string
	Message string
	KeyVals []any
}

func (m *MockTemporalLogger) Debug(msg string, keyvals ...any) {
	m.messages = append(m.messages, LogMessage{Level: "DEBUG", Message: msg, KeyVals: keyvals})
}

func (m *MockTemporalLogger) Info(msg string, keyvals ...any) {
	m.messages = append(m.messages, LogMessage{Level: "INFO", Message: msg, KeyVals: keyvals})
}

func (m *MockTemporalLogger) Warn(msg string, keyvals ...any) {
	m.messages = append(m.messages, LogMessage{Level: "WARN", Message: msg, KeyVals: keyvals})
}

func (m *MockTemporalLogger) Error(msg string, keyvals ...any) {
	m.messages = append(m.messages, LogMessage{Level: "ERROR", Message: msg, KeyVals: keyvals})
}

func Test_TemporalLogHandler(t *testing.T) {
	t.Run("basic logging", func(t *testing.T) {
		mock := &MockTemporalLogger{}
		logger := NewSlogger(mock)

		logger.Info("test message")

		require.Len(t, mock.messages, 1)
		msg := mock.messages[0]
		assert.Equal(t, "INFO", msg.Level)
		assert.Equal(t, "test message", msg.Message)
		assert.Empty(t, msg.KeyVals)
	})

	t.Run("logging with attributes", func(t *testing.T) {
		mock := &MockTemporalLogger{}
		logger := NewSlogger(mock)

		logger.Info("test message", "key1", "value1", "key2", 123)

		require.Len(t, mock.messages, 1)
		msg := mock.messages[0]
		assert.Equal(t, "INFO", msg.Level)
		assert.Equal(t, "test message", msg.Message)
		require.Len(t, msg.KeyVals, 4)
		assert.Equal(t, "key1", msg.KeyVals[0])
		assert.Equal(t, "value1", msg.KeyVals[1])
		assert.Equal(t, "key2", msg.KeyVals[2])
		assert.Equal(t, int64(123), msg.KeyVals[3])
	})

	t.Run("logging with groups", func(t *testing.T) {
		mock := &MockTemporalLogger{}
		logger := NewSlogger(mock)

		groupLogger := logger.WithGroup("group1")
		groupLogger.Info("test message", "key1", "value1")

		require.Len(t, mock.messages, 1)
		msg := mock.messages[0]
		assert.Equal(t, "INFO", msg.Level)
		assert.Equal(t, "test message", msg.Message)
		require.Len(t, msg.KeyVals, 2)
		assert.Equal(t, "group1.key1", msg.KeyVals[0])
		assert.Equal(t, "value1", msg.KeyVals[1])
	})

	t.Run("logging with pre-defined attributes", func(t *testing.T) {
		mock := &MockTemporalLogger{}
		logger := NewSlogger(mock)

		withLogger := logger.With("preset", "value")
		withLogger.Info("test message", "key1", "value1")

		require.Len(t, mock.messages, 1)
		msg := mock.messages[0]
		assert.Equal(t, "INFO", msg.Level)
		assert.Equal(t, "test message", msg.Message)
		require.Len(t, msg.KeyVals, 4)
		assert.Equal(t, "preset", msg.KeyVals[0])
		assert.Equal(t, "value", msg.KeyVals[1])
		assert.Equal(t, "key1", msg.KeyVals[2])
		assert.Equal(t, "value1", msg.KeyVals[3])
	})

	t.Run("all log levels", func(t *testing.T) {
		mock := &MockTemporalLogger{}
		logger := NewSlogger(mock)

		logger.Debug("debug message")
		logger.Info("info message")
		logger.Warn("warn message")
		logger.Error("error message")

		require.Len(t, mock.messages, 4)
		assert.Equal(t, "DEBUG", mock.messages[0].Level)
		assert.Equal(t, "INFO", mock.messages[1].Level)
		assert.Equal(t, "WARN", mock.messages[2].Level)
		assert.Equal(t, "ERROR", mock.messages[3].Level)
	})
}
