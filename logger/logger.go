package logger

import (
	"log"
	"os"

	"github.com/go-web-services/go-web-platform/constants"
	"github.com/go-web-services/go-web-platform/types"
	"github.com/mattn/go-isatty"
)

// ANSI SGR: color only the level label; log body stays default.
const (
	ansiReset = "\x1b[0m"
	ansiDebug = "\x1b[90m" // bright black / gray
	ansiInfo  = "\x1b[32m" // green
	ansiWarn  = "\x1b[33m" // yellow
	ansiError = "\x1b[31m" // red
	ansiFatal = "\x1b[1;31m"
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
	env      types.Environment
	info     *log.Logger
	useColor bool
}

// NewLogger creates a new Logger instance.
// When stdout is a terminal and NO_COLOR is unset, level names are printed with ANSI colors.
func NewLogger(env types.Environment) Logger {
	return &simpleLogger{
		env:  env,
		info: log.New(os.Stdout, "", log.LstdFlags),
		useColor: isatty.IsTerminal(os.Stdout.Fd()) &&
			os.Getenv("NO_COLOR") == "",
	}
}

func (l *simpleLogger) setLevelPrefix(name, colorOpen string) {
	if l.useColor {
		l.info.SetPrefix(colorOpen + name + ansiReset + " ")
		return
	}
	l.info.SetPrefix(name + " ")
}

// Debug logs a message with DEBUG prefix.
func (l *simpleLogger) Debug(args ...interface{}) {
	if l.env == constants.Development || l.env == constants.Staging || l.env == constants.Local {
		l.setLevelPrefix("DEBUG", ansiDebug)
		l.info.Println(args...)
	}
}

// Info logs a message with INFO prefix.
func (l *simpleLogger) Info(args ...interface{}) {
	l.setLevelPrefix("INFO", ansiInfo)
	l.info.Println(args...)
}

// Warn logs a message with WARN prefix.
func (l *simpleLogger) Warn(args ...interface{}) {
	l.setLevelPrefix("WARN", ansiWarn)
	l.info.Println(args...)
}

// Error logs a message with ERROR prefix.
func (l *simpleLogger) Error(args ...interface{}) {
	l.setLevelPrefix("ERROR", ansiError)
	l.info.Println(args...)
}

// Fatal logs a message with FATAL prefix and exits the program.
func (l *simpleLogger) Fatal(args ...interface{}) {
	l.setLevelPrefix("FATAL", ansiFatal)
	l.info.Fatal(args...)
}
