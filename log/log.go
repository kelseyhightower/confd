// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.

/*
Package log provides support for logging to stdout and stderr.

Log entires will be log in the following format:

    timestamp hostname tag[pid]: SEVERITY Message
*/
package log

import (
	"fmt"
	"os"
	"time"
)

// tag represents the application name generating the log message. The tag
// string will appear in all log entires.
var tag string

// Silence non-error messages.
var quiet bool

func init() {
	tag = os.Args[0]
}

// SetTag sets the tag.
func SetTag(t string) {
	tag = t
}

// SetQuiet enables quite mode.
func SetQuiet(q bool) {
	quiet = q
}

// Debug logs a message with severity DEBUG.
func Debug(msg string) {
	write("DEBUG", msg)
}

// Error logs a message with severity ERROR.
func Error(msg string) {
	write("ERROR", msg)
}

// Fatal logs a message with severity ERROR followed by a call to os.Exit().
func Fatal(msg string) {
	write("ERROR", msg)
	os.Exit(1)
}

// Info logs a message with severity INFO.
func Info(msg string) {
	write("INFO", msg)
}

// Notice logs a message with severity NOTICE.
func Notice(msg string) {
	write("NOTICE", msg)
}

// Warning logs a message with severity WARNING.
func Warning(msg string) {
	write("WARNING", msg)
}

// write writes error messages to stderr and all others to stdout.
// Messages are written in the following format:
//     timestamp hostname tag[pid]: SEVERITY Message
func write(level, msg string) {
	var w *os.File
	timestamp := time.Now().Format(time.RFC3339)
	hostname, _ := os.Hostname()
	switch level {
	case "DEBUG", "INFO", "NOTICE", "WARNING":
		if !quiet {
			w = os.Stdout
		}
	case "ERROR":
		w = os.Stderr
	}
	fmt.Fprintf(w, "%s %s %s[%d]: %s %s\n",
		timestamp, hostname, tag, os.Getpid(), level, msg)
}
