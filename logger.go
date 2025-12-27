package dbkit

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.Logger
	sugar  *zap.SugaredLogger
	debug  bool
	atom   zap.AtomicLevel
)

func init() {
	atom = zap.NewAtomicLevelAt(zap.InfoLevel)
}

// InitLogger initializes the logger with the specified level
// level can be: "debug", "info", "warn", "error"
// Outputs to console only
func InitLogger(level string) {
	InitLoggerWithFile(level, "")
}

// InitLoggerWithFile initializes the logger with the specified level and file output
// level can be: "debug", "info", "warn", "error"
// filePath is the log file path (empty string means console output only)
func InitLoggerWithFile(level string, filePath string) {
	// Parse log level
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
		debug = true
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}
	atom.SetLevel(zapLevel)

	// Create encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Determine write syncer
	var writeSyncer zapcore.WriteSyncer
	if filePath != "" {
		// Open log file (append mode, create if not exists)
		file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			// Fallback to console if file creation fails
			writeSyncer = zapcore.AddSync(os.Stdout)
		} else {
			// Write to both console and file
			writeSyncer = zapcore.NewMultiWriteSyncer(
				zapcore.AddSync(os.Stdout),
				zapcore.AddSync(file),
			)
		}
	} else {
		// Console only
		writeSyncer = zapcore.AddSync(os.Stdout)
	}

	// Create core
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		writeSyncer,
		atom,
	)

	// Create logger
	logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	sugar = logger.Sugar()
}

// GetLogger returns the zap logger
func GetLogger() *zap.Logger {
	if logger == nil {
		// Initialize with default info level if not initialized
		InitLogger("info")
	}
	return logger
}

// GetSugarLogger returns the zap sugared logger
func GetSugarLogger() *zap.SugaredLogger {
	if sugar == nil {
		// Initialize with default info level if not initialized
		InitLogger("info")
	}
	return sugar
}

// IsDebugEnabled returns true if debug mode is enabled
func IsDebugEnabled() bool {
	return debug
}

// SetDebugMode enables or disables debug mode
func SetDebugMode(enabled bool) {
	debug = enabled
	if enabled {
		atom.SetLevel(zap.DebugLevel)
	} else {
		atom.SetLevel(zap.InfoLevel)
	}
}

// LogSQL logs SQL statement and parameters in debug mode
func LogSQL(dbName string, sql string, args []interface{}) {
	if debug && logger != nil {
		logger.Debug("SQL executed",
			zap.String("db", dbName),
			zap.String("sql", sql),
			zap.Any("args", args),
		)
	}
}

// LogSQLError logs SQL error
func LogSQLError(dbName string, sql string, args []interface{}, err error) {
	if logger != nil {
		logger.Error("SQL execution failed",
			zap.String("db", dbName),
			zap.String("sql", sql),
			zap.Any("args", args),
			zap.Error(err),
		)
	}
}

// LogInfo logs info message
func LogInfo(msg string, fields ...zap.Field) {
	if logger != nil {
		logger.Info(msg, fields...)
	}
}

// LogWarn logs warning message
func LogWarn(msg string, fields ...zap.Field) {
	if logger != nil {
		logger.Warn(msg, fields...)
	}
}

// LogError logs error message
func LogError(msg string, fields ...zap.Field) {
	if logger != nil {
		logger.Error(msg, fields...)
	}
}

// LogDebug logs debug message
func LogDebug(msg string, fields ...zap.Field) {
	if logger != nil && debug {
		logger.Debug(msg, fields...)
	}
}

// Sync flushes any buffered log entries
func Sync() error {
	if logger != nil {
		return logger.Sync()
	}
	return nil
}
