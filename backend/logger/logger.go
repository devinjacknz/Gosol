package logger

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Logger
	fields logrus.Fields
}

type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	FatalLevel LogLevel = "fatal"
)

// Custom formatter for better debugging
type CustomFormatter struct {
	*logrus.JSONFormatter
}

func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// Add caller info
	if _, file, line, ok := runtime.Caller(8); ok {
		entry.Data["source"] = fmt.Sprintf("%s:%d", trimPath(file), line)
	}

	// Add timestamp in RFC3339Nano format
	entry.Data["@timestamp"] = time.Now().Format(time.RFC3339Nano)

	// Add goroutine ID
	entry.Data["goroutine"] = getGoroutineID()

	// Add memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	entry.Data["memory"] = map[string]interface{}{
		"alloc":      memStats.Alloc,
		"total_alloc": memStats.TotalAlloc,
		"sys":        memStats.Sys,
		"num_gc":     memStats.NumGC,
	}

	return f.JSONFormatter.Format(entry)
}

func trimPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 3 {
		return strings.Join(parts[len(parts)-3:], "/")
	}
	return path
}

func getGoroutineID() uint64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	var id uint64
	fmt.Sscanf(idField, "%d", &id)
	return id
}

// Create a new logger instance
func NewLogger() *Logger {
	log := logrus.New()
	log.SetFormatter(&CustomFormatter{
		&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
		},
	})
	log.SetOutput(os.Stdout)

	// Set log level from environment variable
	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "info"
	}
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	log.SetLevel(logLevel)

	return &Logger{
		Logger: log,
		fields: logrus.Fields{},
	}
}

// WithFields adds fields to the logger
func (l *Logger) WithFields(fields logrus.Fields) *Logger {
	return &Logger{
		Logger: l.Logger,
		fields: fields,
	}
}

// WithContext adds context fields to the logger
func (l *Logger) WithContext(ctx context.Context) *Logger {
	fields := make(logrus.Fields)
	for k, v := range l.fields {
		fields[k] = v
	}

	// Add trace ID if exists
	if traceID := ctx.Value("trace_id"); traceID != nil {
		fields["trace_id"] = traceID
	}

	// Add request ID if exists
	if requestID := ctx.Value("request_id"); requestID != nil {
		fields["request_id"] = requestID
	}

	return &Logger{
		Logger: l.Logger,
		fields: fields,
	}
}

// Debug logs a debug message
func (l *Logger) Debug(args ...interface{}) {
	l.Logger.WithFields(l.fields).Debug(args...)
}

// Debugf logs a debug message with format
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Logger.WithFields(l.fields).Debugf(format, args...)
}

// Info logs an info message
func (l *Logger) Info(args ...interface{}) {
	l.Logger.WithFields(l.fields).Info(args...)
}

// Infof logs an info message with format
func (l *Logger) Infof(format string, args ...interface{}) {
	l.Logger.WithFields(l.fields).Infof(format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(args ...interface{}) {
	l.Logger.WithFields(l.fields).Warn(args...)
}

// Warnf logs a warning message with format
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Logger.WithFields(l.fields).Warnf(format, args...)
}

// Error logs an error message
func (l *Logger) Error(args ...interface{}) {
	l.Logger.WithFields(l.fields).Error(args...)
}

// Errorf logs an error message with format
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Logger.WithFields(l.fields).Errorf(format, args...)
}

// Fatal logs a fatal message
func (l *Logger) Fatal(args ...interface{}) {
	l.Logger.WithFields(l.fields).Fatal(args...)
}

// Fatalf logs a fatal message with format
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.Logger.WithFields(l.fields).Fatalf(format, args...)
}
