package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func init() {
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)
}

// SetLevel sets the logging level for the logger
func SetLevel(level logrus.Level) {
	logger.SetLevel(level)
}

// Info logs information messages
func Info(args ...interface{}) {
	logger.Info(args...)
}

// Infof logs formatted information messages
func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

// Warn logs warning messages
func Warn(args ...interface{}) {
	logger.Warn(args...)
}

// Warnf logs formatted warning messages
func Warnf(format string, args ...interface{}) {
	logger.Warnf(format, args...)
}

// Error logs error messages
func Error(args ...interface{}) {
	logger.Error(args...)
}

// Errorf logs formatted error messages
func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

// Fatal logs fatal messages and exits with status code 1
func Fatal(args ...interface{}) {
	logger.Fatal(args...)
}

// Fatalf logs formatted fatal messages and exits with status code 1
func Fatalf(format string, args ...interface{}) {
	logger.Fatalf(format, args...)
}

// Panic logs panic messages and panics
func Panic(args ...interface{}) {
	logger.Panic(args...)
}

// Panicf logs formatted panic messages and panics
func Panicf(format string, args ...interface{}) {
	logger.Panicf(format, args...)
}

// Add an error as single field to the log entry.  All it does is call
// `WithError` for the given `error`.
func WithError(err error) *logrus.Entry {
	return logger.WithError(err)
}
