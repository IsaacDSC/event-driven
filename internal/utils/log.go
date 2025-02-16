package utils

import (
	"log/slog"
	"os"
)

type Logger struct {
	logger *slog.Logger
	prefix string
}

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
