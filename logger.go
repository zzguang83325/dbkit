package dbkit

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

// LogLevel defines the severity of the log
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l LogLevel) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger interface defines simple behavior for logging with structured fields
type Logger interface {
	// Log records a log entry. fields is optional (can be nil).
	Log(level LogLevel, msg string, fields map[string]interface{})
}

// slogLogger is an adapter for log/slog
type slogLogger struct {
	logger *slog.Logger
}

func (s *slogLogger) Log(level LogLevel, msg string, fields map[string]interface{}) {
	l := s.logger
	if l == nil {
		l = slog.Default()
	}

	// Convert map to slice of key-value pairs for slog with stable order
	var args []interface{}
	if len(fields) > 0 {
		args = make([]interface{}, 0, len(fields)*2)

		// Priority keys to print first in specific order
		priorityKeys := []string{"db", "duration", "sql", "args", "error"}
		processedKeys := make(map[string]bool)

		// 1. Process priority keys first
		for _, k := range priorityKeys {
			if v, ok := fields[k]; ok {
				if k == "args" {
					if slice, ok := v.([]interface{}); ok {
						v = formatValue(slice)
					}
				}
				args = append(args, k, v)
				processedKeys[k] = true
			}
		}

		// 2. Sort remaining keys
		remainingKeys := make([]string, 0, len(fields)-len(processedKeys))
		for k := range fields {
			if !processedKeys[k] {
				remainingKeys = append(remainingKeys, k)
			}
		}
		sort.Strings(remainingKeys)

		// 3. Process remaining keys
		for _, k := range remainingKeys {
			v := fields[k]
			args = append(args, k, v)
		}
	}

	switch level {
	case LevelDebug:
		l.Debug(msg, args...)
	case LevelInfo:
		l.Info(msg, args...)
	case LevelWarn:
		l.Warn(msg, args...)
	case LevelError:
		l.Error(msg, args...)
	}
}

// NewSlogLogger creates a Logger that uses log/slog
func NewSlogLogger(logger *slog.Logger) Logger {
	return &slogLogger{logger: logger}
}

// formatValue formats a log field value
func formatValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("'%s'", val)
	case []interface{}:
		var strs []string
		for _, item := range val {
			if s, ok := item.(string); ok {
				strs = append(strs, fmt.Sprintf("'%s'", s))
			} else {
				strs = append(strs, fmt.Sprintf("%v", item))
			}
		}
		return fmt.Sprintf("[%s]", strings.Join(strs, ", "))
	default:
		return fmt.Sprintf("%v", val)
	}
}

var (
	currentLogger Logger = &slogLogger{logger: nil}
	debug         bool
	re            = regexp.MustCompile(`\s+`)
)

// SetLogger sets the global logger
func SetLogger(l Logger) {
	currentLogger = l
}

// SetDebugMode enables or disables debug mode
func SetDebugMode(enabled bool) {
	debug = enabled
	if enabled {
		// 如果全局 slog 还不支持 Debug 级别，则强制设置一个输出到标准输出的 Debug 级别 slog
		if !slog.Default().Enabled(context.Background(), slog.LevelDebug) {
			slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))
		}
	}
}

// IsDebugEnabled returns true if debug mode is enabled
func IsDebugEnabled() bool {
	return debug
}

// cleanSQL removes newlines, tabs and multiple spaces from SQL string
func cleanSQL(sql string) string {
	return strings.TrimSpace(re.ReplaceAllString(sql, " "))
}

// LogSQL logs SQL statement, parameters and execution time in debug mode
func LogSQL(dbName string, sql string, args []interface{}, duration time.Duration) {
	if debug {
		fields := map[string]interface{}{
			"db":       dbName,
			"sql":      cleanSQL(sql),
			"duration": duration.String(),
		}
		if len(args) > 0 {
			fields["args"] = args
		}
		currentLogger.Log(LevelDebug, "SQL log", fields)
	}
}

// LogSQLError logs SQL error with execution time
func LogSQLError(dbName string, sql string, args []interface{}, duration time.Duration, err error) {
	fields := map[string]interface{}{
		"db":       dbName,
		"sql":      cleanSQL(sql),
		"duration": duration.String(),
		"error":    err.Error(),
	}
	if len(args) > 0 {
		fields["args"] = args
	}
	currentLogger.Log(LevelError, "SQL failed log", fields)
}

// LogInfo logs info message
func LogInfo(msg string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	currentLogger.Log(LevelInfo, msg, f)
}

// LogWarn logs warning message
func LogWarn(msg string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	currentLogger.Log(LevelWarn, msg, f)
}

// LogError logs error message
func LogError(msg string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	currentLogger.Log(LevelError, msg, f)
}

// LogDebug logs debug message
func LogDebug(msg string, fields ...map[string]interface{}) {
	if debug {
		var f map[string]interface{}
		if len(fields) > 0 {
			f = fields[0]
		}
		currentLogger.Log(LevelDebug, msg, f)
	}
}

// Sync flushes any buffered log entries
func Sync() {
	if s, ok := currentLogger.(interface{ Sync() error }); ok {
		_ = s.Sync()
	}
}

// InitLogger initializes the logger with a specific level to stdout
// InitLogger initializes the global slog logger with a specific level to console
func InitLogger(level string) {
	// Determine log level
	slogLevel := slog.LevelInfo
	if strings.EqualFold(level, "debug") {
		slogLevel = slog.LevelDebug
		SetDebugMode(true)
	} else if strings.EqualFold(level, "warn") {
		slogLevel = slog.LevelWarn
	} else if strings.EqualFold(level, "error") {
		slogLevel = slog.LevelError
	}

	// Set global slog default with TextHandler to stdout
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slogLevel,
	})
	slog.SetDefault(slog.New(handler))

	// Reset currentLogger to use the new global default
	SetLogger(&slogLogger{logger: nil})
}

// InitLoggerWithFile initializes the logger to both console and a file using slog
func InitLoggerWithFile(level string, filePath string) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dbkit: Failed to open log file: %v\n", err)
		return
	}

	// Determine log level
	slogLevel := slog.LevelInfo
	if strings.EqualFold(level, "debug") {
		slogLevel = slog.LevelDebug
		SetDebugMode(true)
	} else if strings.EqualFold(level, "warn") {
		slogLevel = slog.LevelWarn
	} else if strings.EqualFold(level, "error") {
		slogLevel = slog.LevelError
	}

	// Create a multi-writer for both console and file
	multiWriter := io.MultiWriter(os.Stdout, file)

	// Set global slog default with TextHandler
	handler := slog.NewTextHandler(multiWriter, &slog.HandlerOptions{
		Level: slogLevel,
	})
	slog.SetDefault(slog.New(handler))

	// Reset currentLogger to use the new global default
	SetLogger(&slogLogger{logger: nil})
}
