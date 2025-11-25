package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is an interface that abstracts logging functionality.
// This allows for easy testing and potential swapping of logging implementations.
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	With(fields ...Field) Logger
}

// Field represents a key-value pair for structured logging.
type Field = zap.Field

// zapLogger wraps zap.Logger to implement our Logger interface.
type zapLogger struct {
	logger *zap.Logger
}

// Debug logs a debug-level message.
func (l *zapLogger) Debug(msg string, fields ...Field) {
	l.logger.Debug(msg, fields...)
}

// Info logs an info-level message.
func (l *zapLogger) Info(msg string, fields ...Field) {
	l.logger.Info(msg, fields...)
}

// Warn logs a warn-level message.
func (l *zapLogger) Warn(msg string, fields ...Field) {
	l.logger.Warn(msg, fields...)
}

// Error logs an error-level message.
func (l *zapLogger) Error(msg string, fields ...Field) {
	l.logger.Error(msg, fields...)
}

// With creates a new logger with additional fields.
func (l *zapLogger) With(fields ...Field) Logger {
	return &zapLogger{logger: l.logger.With(fields...)}
}

// NewLogger creates a new structured logger based on the provided log level.
// The logger uses zap's production configuration with configurable level.
func NewLogger(level string) (Logger, error) {
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapLevel)
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &zapLogger{logger: logger}, nil
}

// Convenience functions for creating common field types.
// These wrap zap's field constructors for easier use.

// String creates a string field.
func String(key, value string) Field {
	return zap.String(key, value)
}

// Int creates an integer field.
func Int(key string, value int) Field {
	return zap.Int(key, value)
}

// Int64 creates an int64 field.
func Int64(key string, value int64) Field {
	return zap.Int64(key, value)
}

// ErrorField creates an error field.
func ErrorField(err error) Field {
	return zap.Error(err)
}

// Duration creates a duration field.
func Duration(key string, value int64) Field {
	return zap.Int64(key, value)
}

