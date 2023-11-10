package log

import (
	"os"

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
// If fileName is non-empty, it pipes logs to file and stdout.
// if filename is empty, it pipes logs only to stdout.
func NewLogger(isProduction bool, fileName string, logLevelStr string) (Logger, error) {
	loggerConfig := zap.NewProductionConfig()
	if !isProduction {
		loggerConfig = zap.NewDevelopmentConfig()
	}

	logLevel := zap.InfoLevel
	isDebugLevel := logLevelStr == "debug"
	if isDebugLevel {
		logLevel = zap.DebugLevel
	}

	loggerConfig.Level.SetLevel(logLevel)

	consoleEncoder := zapcore.NewConsoleEncoder(loggerConfig.EncoderConfig)

	// Configure piping to stdout and to fileName if it is non-empty
	var core zapcore.Core
	if fileName != "" {
		fileEncoder := zapcore.NewJSONEncoder(loggerConfig.EncoderConfig)

		f, err := os.Create(fileName)
		if err != nil {
			return nil, err
		}

		core = zapcore.NewTee(
			zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.InfoLevel),
			zapcore.NewCore(fileEncoder, zapcore.AddSync(f), zapcore.InfoLevel),
		)
	} else {
		core = zapcore.NewTee(
			zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.InfoLevel),
		)
	}

	logger := zap.New(core)

	logger.Info("log level", zap.Bool("is_debug", isDebugLevel), zap.String("log_level", loggerConfig.Level.String()))

	return logger, nil
}
