package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
}

type NoOpLogger struct{}

// Debug implements Logger.
func (*NoOpLogger) Debug(msg string, fields ...zapcore.Field) {
	// no-op
}

// Error implements Logger.
func (*NoOpLogger) Error(msg string, fields ...zapcore.Field) {
	// no-op
}

// Info implements Logger.
func (*NoOpLogger) Info(msg string, fields ...zapcore.Field) {
	// no-op
}

// Warn implements Logger.
func (*NoOpLogger) Warn(msg string, fields ...zapcore.Field) {
	// no-op
}

var _ Logger = (*NoOpLogger)(nil)

type loggerImpl struct {
	zapLogger zap.Logger
}

var _ Logger = (*loggerImpl)(nil)

// Debug implements Logger.
func (l *loggerImpl) Debug(msg string, fields ...zapcore.Field) {
	l.zapLogger.Debug(msg, fields...)
}

// Error implements Logger.
func (l *loggerImpl) Error(msg string, fields ...zapcore.Field) {
	l.zapLogger.Error(msg, fields...)
}

// Info implements Logger.
func (l *loggerImpl) Info(msg string, fields ...zapcore.Field) {
	l.zapLogger.Info(msg, fields...)
}

// Warn implements Logger.
func (l *loggerImpl) Warn(msg string, fields ...zapcore.Field) {
	l.zapLogger.Warn(msg, fields...)
}

// NewLogger creates a new logger.
func NewLogger(isProduction bool) (Logger, error) {
	// logger
	// TODO: figure out logging to file
	loggerConfig := zap.NewProductionConfig()
	if !isProduction {
		loggerConfig = zap.NewDevelopmentConfig()
	}

	logger, err := loggerConfig.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}
