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

type loggerImpl struct {
	zapLogger *zap.Logger
}

var _ Logger = (*loggerImpl)(nil)

// Debug implements Logger.
func (l *loggerImpl) Debug(msg string, fields ...zapcore.Field) {
	if l.zapLogger != nil {
		l.zapLogger.Debug(msg, fields...)
	}
}

// Error implements Logger.
func (l *loggerImpl) Error(msg string, fields ...zapcore.Field) {
	if l.zapLogger != nil {
		l.zapLogger.Error(msg, fields...)
	}
}

// Info implements Logger.
func (l *loggerImpl) Info(msg string, fields ...zapcore.Field) {
	if l.zapLogger != nil {
		l.zapLogger.Info(msg, fields...)
	}
}

// Warn implements Logger.
func (l *loggerImpl) Warn(msg string, fields ...zapcore.Field) {
	if l.zapLogger != nil {
		l.zapLogger.Warn(msg, fields...)
	}
}

// NewLogger creates a new logger.
func NewLogger() (Logger, error) {
	// logger
	// TODO: figure out logging to file
	loggerConfig := zap.NewProductionConfig()
	logger, err := loggerConfig.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}
