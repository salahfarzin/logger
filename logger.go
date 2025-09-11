package logger

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	envLogLevel = "LOG_LEVEL"
	envLogPath  = "LOG_PATH"
)

var (
	globalLogger *zap.Logger
	once         sync.Once
)

// Init initializes the global zap logger. Call this early in main().
func Init() {
	once.Do(func() {
		cfg := zap.Config{
			OutputPaths: []string{getOutput()},
			Level:       zap.NewAtomicLevelAt(getLevel()),
			Encoding:    "json",
			EncoderConfig: zapcore.EncoderConfig{
				LevelKey:     "level",
				TimeKey:      "time",
				MessageKey:   "msg",
				EncodeTime:   zapcore.ISO8601TimeEncoder,
				EncodeLevel:  zapcore.LowercaseLevelEncoder,
				EncodeCaller: zapcore.ShortCallerEncoder,
			},
		}
		var err error
		globalLogger, err = cfg.Build()
		if err != nil {
			panic(fmt.Sprintf("failed to initialize logger: %v", err))
		}
	})
}

// Get returns the global zap.Logger. Panics if not initialized.
func Get() *zap.Logger {
	if globalLogger == nil {
		panic("Logger not initialized, call logger.Init() first")
	}
	return globalLogger
}

// With returns a child logger with additional context fields.
func With(fields ...zap.Field) *zap.Logger {
	return Get().With(fields...)
}

// Info logs an info message with optional fields.
func Info(msg string, fields ...zap.Field) {
	Get().Info(msg, fields...)
}

// Error logs an error message with optional fields.
func Error(msg string, err error, fields ...zap.Field) {
	allFields := append(fields, zap.NamedError("error", err))
	Get().Error(msg, allFields...)
}

// Debug logs a debug message with optional fields.
func Debug(msg string, fields ...zap.Field) {
	Get().Debug(msg, fields...)
}

// Sync flushes any buffered log entries. Call this on shutdown.
func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}

// getLevel parses the log level from environment variable.
func getLevel() zapcore.Level {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(envLogLevel))) {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	default:
		return zap.InfoLevel
	}
}

// getOutput parses the log output from environment variable.
func getOutput() string {
	output := strings.TrimSpace(os.Getenv(envLogPath))
	if output == "" {
		return "stdout"
	}
	return output
}

// Printf and Print for compatibility with std log interface (optional)
func Printf(format string, v ...interface{}) {
	Info(fmt.Sprintf(format, v...))
}

func Print(v ...interface{}) {
	Info(fmt.Sprint(v...))
}
