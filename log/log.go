package log

import (
	"fmt"
	"os"
	"time"
)

var tag string

func init() {
	tag = os.Args[0]
}

func SetTag(t string) {
	tag = t
}

func Debug(msg string) {
	write("DEBUG", msg)
}

func Error(msg string) {
	write("ERROR", msg)
}

func Fatal(msg string) {
	write("ERROR", msg)
	os.Exit(1)
}

func Info(msg string) {
	write("INFO", msg)
}

func Notice(msg string) {
	write("NOTICE", msg)
}

func Warning(msg string) {
	write("WARNING", msg)
}

func write(level, msg string) {
	var w *os.File
	timestamp := time.Now().Format(time.RFC3339)
	hostname, _ := os.Hostname()
	switch level {
	case "DEBUG", "INFO", "NOTICE", "WARNING":
		w = os.Stdout
	case "ERROR":
		w = os.Stderr
	}
	fmt.Fprintf(w, "%s %s %s[%d]: %s %s\n",
		timestamp, hostname, tag, os.Getpid(), level, msg)
}
