package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"
)

var logTimeFormat = "15:04:05"
var logFileHandle *os.File

// openLogFile - Create the log file, creating the logs directory first if necessary.
func openLogFile(dirPath string, logFileName string) {
	var err error

	// Establish the logs directory.
	err = isDirectory(dirPath)
	if err != nil {
		fatalErr(err, "OpenLogFilePath: isDirectory(%s) failed", dirPath)
	}

	// If a previous log is still open, close it.
	closeLogFile()

	// Create the log file.
	fullPath := filepath.Join(dirPath, logFileName)
	logFileHandle, err = os.Create(fullPath)
	if err != nil {
		fatalErr(err, "OpenLogFilePath: os.Create(%s) failed", fullPath)
	}
}

// closeLogFile - Close the log file.
func closeLogFile() {
	if logFileHandle != nil {
		err := logFileHandle.Close()
		if err != nil {
			fatalErr(err, "CloseLogFile: logFileHandle.Close() failed")
		}
	}
	logFileHandle = nil
}

// logger - Time-stamp log function.
func logger(format string, arg ...any) {
	now := time.Now()
	msg := fmt.Sprintf(format, arg...)
	_, _ = fmt.Fprintf(os.Stdout, "%s %s\n", now.Format(logTimeFormat), msg)
	_ = os.Stdout.Sync()
	writeOutputText(logFileHandle, msg)
}

// loggerSkip - Time-stamp log function with a preceding newline for readability.
func loggerSkip(format string, arg ...any) {
	now := time.Now()
	msg := fmt.Sprintf(format, arg...)
	_, _ = fmt.Fprintf(os.Stdout, "\n%s %s\n", now.Format(logTimeFormat), msg)
	_ = os.Stdout.Sync()
	writeOutputText(logFileHandle, msg)
}

// logWarning - Log a warning.
func logWarning(format string, arg ...any) {
	msg := fmt.Sprintf(format, arg...)
	logger("*** WARNING :: %s", msg)
}

// logError - Log an error.
func logError(format string, arg ...any) {
	msg := fmt.Sprintf(format, arg...)
	logger("*** ERROR :: %s", msg)
}

// fatalText - Log a fatal error message and exit.
func fatalText(format string, arg ...any) {
	text := fmt.Sprintf(format, arg...)
	logger("*** FATAL :: %s", text)
	debug.PrintStack()
	DBClose()
	os.Exit(1)
}

// fatalErr - Log a fatal error message with error text and exit.
func fatalErr(err error, format string, arg ...any) {
	msg := fmt.Sprintf(format, arg...)
	logger("*** FATAL :: %s\n\terr: %s", msg, err.Error())
	debug.PrintStack()
	DBClose()
	os.Exit(1)
}
