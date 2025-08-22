package logger

import (
	"log"
	"os"

	"github.com/Lomank123/go-web-platform/constants"
	"github.com/Lomank123/go-web-platform/types"
)

// Logger is the interface that wraps the basic logging methods.
type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
}

type simpleLogger struct {
	env  types.Environment
	info *log.Logger
}

// NewLogger creates a new Logger instance.
func NewLogger(env types.Environment) Logger {
	return &simpleLogger{
		env:  env,
		info: log.New(os.Stdout, "", log.LstdFlags),
	}
}

// Debug logs a message with DEBUG prefix.
func (l *simpleLogger) Debug(args ...interface{}) {
	if l.env == constants.Development || l.env == constants.Staging || l.env == constants.Local {
		l.info.SetPrefix("DEBUG ")
		l.info.Println(args...)
	}
}

// Info logs a message with INFO prefix.
func (l *simpleLogger) Info(args ...interface{}) {
	l.info.SetPrefix("INFO ")
	l.info.Println(args...)
}

// Warn logs a message with WARN prefix.
func (l *simpleLogger) Warn(args ...interface{}) {
	l.info.SetPrefix("WARN ")
	l.info.Println(args...)
}

// Error logs a message with ERROR prefix.
func (l *simpleLogger) Error(args ...interface{}) {
	l.info.SetPrefix("ERROR ")
	l.info.Println(args...)
}

// Fatal logs a message with FATAL prefix and exits the program.
func (l *simpleLogger) Fatal(args ...interface{}) {
	l.info.SetPrefix("FATAL ")
	l.info.Fatal(args...)
}
