package utils

import (
	"github.com/IsaacDSC/event-driven/types"
	"log/slog"
	"os"
)

type Logger struct {
	logger *slog.Logger
	prefix string
}

type KeyLogger string

func (k KeyLogger) String() string {
	return string(k)
}

const (
	KeyLogError      KeyLogger = "error"
	KeyAsynqTypeTask KeyLogger = "task.name"
	KeyAsynqRetry    KeyLogger = "task.retry"
	KeyAsynqTaskID   KeyLogger = "task.id"
	KeyAsynqElapsed  KeyLogger = "task.elapsed"
)

var _ types.Logger = (*Logger)(nil)

func NewLogger(prefix string) *Logger {
	return &Logger{
		logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})),
		prefix: prefix,
	}
}

func (l *Logger) Info(msg string, keysAndValues ...interface{}) {
	l.logger.Info(l.prefix+msg, keysAndValues...)
}

func (l *Logger) Error(msg string, keysAndValues ...interface{}) {
	l.logger.Error(l.prefix+msg, keysAndValues...)
}

func (l *Logger) Debug(msg string, keysAndValues ...interface{}) {
	l.logger.Debug(l.prefix+msg, keysAndValues...)
}

func (l *Logger) Warn(msg string, keysAndValues ...interface{}) {
	l.logger.Warn(l.prefix+msg, keysAndValues...)
}
