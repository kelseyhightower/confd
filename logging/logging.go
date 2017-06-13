package logging

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/logutils"
)

var validLevels = []logutils.LogLevel{"TRACE", "DEBUG", "INFO", "WARN", "ERROR"}

// LogOutput determines where we should send logs (if anywhere) and the log level.
func LogOutput(level string) io.Writer {
	logLevel := "TRACE"
	if isValidLogLevel(level) {
		// allow following for better ux: info, Info or INFO
		logLevel = strings.ToUpper(level)
	} else {
		log.Printf("[WARN] Invalid log level: %q. Defaulting to level: TRACE. Valid levels are: %+v",
			level, validLevels)
	}

	logOutput := &logutils.LevelFilter{
		Levels:   validLevels,
		MinLevel: logutils.LogLevel(logLevel),
		Writer:   os.Stderr,
	}

	return logOutput
}

// SetLevel checks for a log destination with LogOutput, and calls
// log.SetOutput with the result. If LogOutput returns nil, SetOutput uses
// ioutil.Discard. Any error from LogOutout is fatal.
func SetLevel(level string) {
	out := LogOutput(level)
	if out == nil {
		out = ioutil.Discard
	}

	log.SetOutput(out)
}

func isValidLogLevel(level string) bool {
	for _, l := range validLevels {
		if strings.ToUpper(level) == string(l) {
			return true
		}
	}

	return false
}
