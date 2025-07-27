package logger

import (
	"fmt"
	"io"
	"os"
	"time"
)

// Level represents log levels
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

// String returns string representation of log level
func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Color returns ANSI color code for log level
func (l Level) Color() string {
	switch l {
	case DEBUG:
		return "\033[36m" // Cyan
	case INFO:
		return "\033[32m" // Green
	case WARN:
		return "\033[33m" // Yellow
	case ERROR:
		return "\033[31m" // Red
	case FATAL:
		return "\033[35m" // Magenta
	default:
		return "\033[0m" // Reset
	}
}

// Logger provides structured logging
type Logger struct {
	level      Level
	output     io.Writer
	timeFormat string
	colored    bool
	prefix     string
}

// New creates a new logger instance
func New(level Level, output io.Writer) *Logger {
	return &Logger{
		level:      level,
		output:     output,
		timeFormat: "2006-01-02 15:04:05",
		colored:    isTerminal(output),
	}
}

// Default creates a default logger to stdout
func Default() *Logger {
	return New(INFO, os.Stdout)
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level Level) {
	l.level = level
}

// SetPrefix sets a prefix for all log messages
func (l *Logger) SetPrefix(prefix string) {
	l.prefix = prefix
}

// EnableColors enables/disables colored output
func (l *Logger) EnableColors(enabled bool) {
	l.colored = enabled
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(FATAL, format, args...)
	os.Exit(1)
}

// AttackStarted logs attack initiation
func (l *Logger) AttackStarted(algorithm, attackType string, threads int) {
	l.Info("🎯 Attack started: %s algorithm, %s mode, %d threads", algorithm, attackType, threads)
}

// AttackCompleted logs attack completion
func (l *Logger) AttackCompleted(success bool, attempts uint64, duration time.Duration, secret string) {
	if success {
		l.Info("✅ Attack successful: found secret '%s' after %d attempts in %v", secret, attempts, duration)
	} else {
		l.Info("❌ Attack completed: no secret found after %d attempts in %v", attempts, duration)
	}
}

// TokenAnalyzed logs JWT token analysis
func (l *Logger) TokenAnalyzed(algorithm string, valid bool) {
	if valid {
		l.Info("🔍 Token analyzed: valid JWT with %s algorithm", algorithm)
	} else {
		l.Warn("⚠️  Token analyzed: invalid or malformed JWT")
	}
}

// WebServerStarted logs web server startup
func (l *Logger) WebServerStarted(port int) {
	l.Info("🌐 Web server started on port %d", port)
}

// ProgressUpdate logs attack progress
func (l *Logger) ProgressUpdate(attempts uint64, rate float64, eta time.Duration) {
	l.Debug("Progress: %d attempts, %.1f/s, ETA: %v", attempts, rate, eta)
}

// log handles the actual logging with formatting
func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.level {
		return
	}
	
	timestamp := time.Now().Format(l.timeFormat)
	message := fmt.Sprintf(format, args...)
	
	var levelStr string
	if l.colored {
		levelStr = fmt.Sprintf("%s%-5s\033[0m", level.Color(), level.String())
	} else {
		levelStr = fmt.Sprintf("%-5s", level.String())
	}
	
	var logLine string
	if l.prefix != "" {
		logLine = fmt.Sprintf("[%s] %s [%s] %s\n", timestamp, levelStr, l.prefix, message)
	} else {
		logLine = fmt.Sprintf("[%s] %s %s\n", timestamp, levelStr, message)
	}
	
	l.output.Write([]byte(logLine))
}

// isTerminal checks if the output is a terminal
func isTerminal(w io.Writer) bool {
	if w == os.Stdout || w == os.Stderr {
		return true
	}
	return false
}

// Global logger instance
var defaultLogger = Default()

// Global logging functions
func Debug(format string, args ...interface{}) {
	defaultLogger.Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	defaultLogger.Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	defaultLogger.Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	defaultLogger.Error(format, args...)
}

func Fatal(format string, args ...interface{}) {
	defaultLogger.Fatal(format, args...)
}

func SetLevel(level Level) {
	defaultLogger.SetLevel(level)
}

func SetPrefix(prefix string) {
	defaultLogger.SetPrefix(prefix)
}