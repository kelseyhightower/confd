/*
Package log provides support for logging to stdout and stderr.

Log entries will be logged in the following format:

    timestamp hostname tag[pid]: SEVERITY Message
*/
package log

import (
	"fmt"
	"os"

	hclog "github.com/hashicorp/go-hclog"
)

var log hclog.Logger

func init() {
	log = hclog.New(&hclog.LoggerOptions{
		Output: os.Stderr,
		Level:  hclog.Trace,
	})
	if os.Args[1] == "internal-plugin" {
		log = hclog.New(&hclog.LoggerOptions{
			Output:     os.Stderr,
			Level:      hclog.Trace,
			JSONFormat: true,
		})
	}
}

func GetLogger() *hclog.Logger {
	return &log
}

// SetLevel sets the log level. Valid levels are panic, fatal, error, warn, info and debug.
func SetLevel(level string) {
	log = hclog.New(&hclog.LoggerOptions{
		Output: os.Stderr,
		Level:  hclog.LevelFromString(level),
	})
	if os.Args[1] == "internal-plugin" {
		log = hclog.New(&hclog.LoggerOptions{
			Output:     os.Stderr,
			Level:      hclog.LevelFromString(level),
			JSONFormat: true,
		})
	}
}

// Debug logs a message with severity DEBUG.
func Debug(format string, v ...interface{}) {
	log.Debug(fmt.Sprintf(format, v...))
}

// Error logs a message with severity ERROR.
func Error(format string, v ...interface{}) {
	log.Error(fmt.Sprintf(format, v...))
}

// Fatal logs a message with severity ERROR followed by a call to os.Exit().
func Fatal(format string, v ...interface{}) {
	log.Error(fmt.Sprintf(format, v...))
	os.Exit(1)
}

// Info logs a message with severity INFO.
func Info(format string, v ...interface{}) {
	log.Info(fmt.Sprintf(format, v...))
}

// Warning logs a message with severity WARNING.
func Warning(format string, v ...interface{}) {
	log.Warn(fmt.Sprintf(format, v...))
}
