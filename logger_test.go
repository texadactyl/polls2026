package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoggerFunctions(t *testing.T) {
	tempDir := t.TempDir()
	logName := "test_log.log"

	openLogFile(tempDir, logName)
	defer closeLogFile()

	logger("test logger line: %s", "hello")
	loggerSkip("test loggerSkip line: %d", 42)
	logWarning("test warning: %s", "caution")
	logError("test error: %s", "problem")

	closeLogFile()

	logContent, err := os.ReadFile(filepath.Join(tempDir, logName))
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	contentStr := string(logContent)
	if !strings.Contains(contentStr, "test logger line: hello") {
		t.Fatalf("missing logger message in log file")
	}
	if !strings.Contains(contentStr, "test loggerSkip line: 42") {
		t.Fatalf("missing loggerSkip message in log file")
	}
	if !strings.Contains(contentStr, "*** WARNING :: test warning: caution") {
		t.Fatalf("missing warning message in log file")
	}
	if !strings.Contains(contentStr, "*** ERROR :: test error: problem") {
		t.Fatalf("missing error message in log file")
	}
}
