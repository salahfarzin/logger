package logger

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	envLogLevel = "LOG_LEVEL"
	envLogPath  = "LOG_PATH"
)

var (
	globalLogger *zap.Logger
	globalSyncer *dailyRotateSyncer
	once         sync.Once
	logDir       = "logs" // directory used by the daily rotator; override with SetLogDir before Init
)

// SetLogDir sets the directory where daily-rotating log files are written.
// Must be called before Init. Has no effect after Init has been called.
func SetLogDir(dir string) {
	logDir = dir
}

// dailyRotateSyncer is a zapcore.WriteSyncer that rotates the log file at
// midnight. On the first Write after the calendar date changes the old file is
// closed and a new YYYY-MM-DD.log is opened in the same directory.
type dailyRotateSyncer struct {
	mu      sync.Mutex
	dir     string
	current *os.File
	date    string // "YYYY-MM-DD" of the currently open file
}

func newDailyRotateSyncer(dir string) (*dailyRotateSyncer, error) {
	s := &dailyRotateSyncer{dir: dir}
	if err := s.rotateLocked(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *dailyRotateSyncer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if todayDate() != s.date {
		if err := s.rotateLocked(); err != nil {
			return 0, err
		}
	}
	return s.current.Write(p)
}

func (s *dailyRotateSyncer) Sync() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current == nil {
		return nil
	}
	return s.current.Sync()
}

func (s *dailyRotateSyncer) close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current == nil {
		return nil
	}
	err := s.current.Close()
	s.current = nil
	return err
}

// rotateLocked opens a new date-named file. Must be called with s.mu held.
func (s *dailyRotateSyncer) rotateLocked() error {
	if s.current != nil {
		_ = s.current.Sync()
		_ = s.current.Close()
		s.current = nil
	}
	day := todayDate()
	path := fmt.Sprintf("%s/%s.log", s.dir, day)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
	if err != nil {
		return fmt.Errorf("logger: open %s: %w", path, err)
	}
	s.current = f
	s.date = day
	return nil
}

func todayDate() string {
	t := time.Now()
	return fmt.Sprintf("%04d-%02d-%02d", t.Year(), t.Month(), t.Day())
}

// mergeEncoderConfig merges user-provided EncoderConfig into the default.
func mergeEncoderConfig(def, user zapcore.EncoderConfig) zapcore.EncoderConfig {
	if user.LevelKey != "" {
		def.LevelKey = user.LevelKey
	}
	if user.TimeKey != "" {
		def.TimeKey = user.TimeKey
	}
	if user.MessageKey != "" {
		def.MessageKey = user.MessageKey
	}
	if user.EncodeTime != nil {
		def.EncodeTime = user.EncodeTime
	}
	if user.EncodeLevel != nil {
		def.EncodeLevel = user.EncodeLevel
	}
	if user.EncodeCaller != nil {
		def.EncodeCaller = user.EncodeCaller
	}
	return def
}

// Init initializes the global zap logger. Pass a *zap.Config to override
// encoding, level, or encoder fields. Output paths in the config are ignored —
// use the LOG_PATH env var to write to a fixed file instead of rotating files.
func Init(cnf ...*zap.Config) {
	var userCfg *zap.Config
	if len(cnf) > 0 {
		userCfg = cnf[0]
	}

	once.Do(func() {
		encoderCfg := zapcore.EncoderConfig{
			LevelKey:     "level",
			TimeKey:      "time",
			MessageKey:   "msg",
			EncodeTime:   zapcore.ISO8601TimeEncoder,
			EncodeLevel:  zapcore.LowercaseLevelEncoder,
			EncodeCaller: zapcore.ShortCallerEncoder,
		}

		level := getLevel()
		encoding := "json"

		if userCfg != nil {
			encoderCfg = mergeEncoderConfig(encoderCfg, userCfg.EncoderConfig)
			if userCfg.Level.Level() != 0 {
				level = userCfg.Level.Level()
			}
			if userCfg.Encoding != "" {
				encoding = userCfg.Encoding
			}
		}

		var enc zapcore.Encoder
		if encoding == "console" {
			enc = zapcore.NewConsoleEncoder(encoderCfg)
		} else {
			enc = zapcore.NewJSONEncoder(encoderCfg)
		}

		ws := buildWriteSyncer()

		core := zapcore.NewCore(enc, ws, zap.NewAtomicLevelAt(level))
		globalLogger = zap.New(core, zap.AddCaller())
	})
}

// buildWriteSyncer returns a WriteSyncer based on the LOG_PATH env var.
// If LOG_PATH is set, it is always treated as a rotation directory.
// Otherwise it creates a dailyRotateSyncer writing to logDir/YYYY-MM-DD.log.
func buildWriteSyncer() zapcore.WriteSyncer {
	if staticPath := strings.TrimSpace(os.Getenv(envLogPath)); staticPath != "" {
		logDir = staticPath
	}

	if err := os.MkdirAll(logDir, 0755); err != nil {
		return zapcore.AddSync(os.Stdout)
	}
	s, err := newDailyRotateSyncer(logDir)
	if err != nil {
		panic(fmt.Sprintf("logger: create daily rotator: %v", err))
	}
	globalSyncer = s
	return s
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

// Close flushes and closes the underlying log file. Call this on shutdown
// alongside or instead of Sync.
func Close() error {
	_ = Sync()
	if globalSyncer != nil {
		return globalSyncer.close()
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

// Printf and Print for compatibility with std log interface.
func Printf(format string, v ...any) {
	Info(fmt.Sprintf(format, v...))
}

func Print(v ...any) {
	Info(fmt.Sprint(v...))
}
