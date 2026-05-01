package logger

import (
	"log"
	"os"

	"github.com/go-web-services/go-web-platform/constants"
	"github.com/go-web-services/go-web-platform/types"
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

// NewLogger creates a new Logger instance that writes plain-text, level-prefixed lines to stdout.
func NewLogger(env types.Environment) Logger {
	return &simpleLogger{
		env:  env,
		info: log.New(os.Stdout, "", log.LstdFlags),
	}
}

func (l *simpleLogger) setLevelPrefix(name string) {
	l.info.SetPrefix("[" + name + "] ")
}

// Debug logs a message with [DEBUG] prefix.
func (l *simpleLogger) Debug(args ...interface{}) {
	if l.env == constants.Development || l.env == constants.Staging || l.env == constants.Local {
		l.setLevelPrefix("DEBUG")
		l.info.Println(args...)
	}
}

// Info logs a message with [INFO] prefix.
func (l *simpleLogger) Info(args ...interface{}) {
	l.setLevelPrefix("INFO")
	l.info.Println(args...)
}

// Warn logs a message with [WARN] prefix.
func (l *simpleLogger) Warn(args ...interface{}) {
	l.setLevelPrefix("WARN")
	l.info.Println(args...)
}

// Error logs a message with [ERROR] prefix.
func (l *simpleLogger) Error(args ...interface{}) {
	l.setLevelPrefix("ERROR")
	l.info.Println(args...)
}

// Fatal logs a message with [FATAL] prefix and exits the program.
func (l *simpleLogger) Fatal(args ...interface{}) {
	l.setLevelPrefix("FATAL")
	l.info.Fatal(args...)
}
