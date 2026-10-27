package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// captureOutput captures stdout output generated while executing fn.
func captureOutput(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}
	os.Stdout = w

	outChan := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outChan <- buf.String()
	}()

	fn()

	_ = w.Close()
	os.Stdout = oldStdout
	out := <-outChan
	_ = r.Close()

	return out
}

func setupTestDB(t *testing.T) {
	t.Helper()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	absDbPath, err := filepath.Abs(dbPath)
	if err != nil {
		t.Fatalf("filepath.Abs failed for %s: %v", dbPath, err)
	}

	dsn := fmt.Sprintf("%s?_pragma=busy_timeout(5000)", absDbPath)
	handle, err := sql.Open(driverDatabase, dsn)
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}

	prevHandle := dbHandle
	prevIsOpen := dbIsOpen
	t.Cleanup(func() {
		DBClose()
		dbHandle = prevHandle
		dbIsOpen = prevIsOpen
		if err := os.Remove(absDbPath); err != nil && !os.IsNotExist(err) {
			t.Errorf("failed to remove database file (%s): %v", absDbPath, err)
		}
	})

	dbHandle = handle
	dbIsOpen = true

	if err := sqlFunc(fmtSqlCreateHistory); err != nil {
		t.Fatalf("failed to create history table: %v", err)
	}
}

func TestReportByState_NoRecords(t *testing.T) {
	setupTestDB(t)

	out := captureOutput(t, func() {
		reportByState("NV")
	})

	expected := "No history records were found for state NV"
	if !strings.Contains(out, expected) {
		t.Fatalf("expected output to contain %q, got:\n%s", expected, out)
	}
}

func TestReportByState_WithRecords(t *testing.T) {
	setupTestDB(t)

	// Insert test records
	// CA record 1 in October (endMonth=10, endDay=15)
	err := DBInsert("hash_ca_1", "2026-10-15_10:00:00", "CA", 55.0, 40.0, 3.0, 2.0, 10, 1, 10, 15, "PollsterOct1")
	if err != nil {
		t.Fatalf("DBInsert failed: %v", err)
	}

	// CA record 2 in October (endMonth=10, endDay=25)
	err = DBInsert("hash_ca_2", "2026-10-25_10:00:00", "CA", 56.0, 39.0, 3.0, 2.0, 10, 16, 10, 25, "PollsterOct2")
	if err != nil {
		t.Fatalf("DBInsert failed: %v", err)
	}

	// CA record 3 in September (endMonth=9, should be filtered out by endMonth < 10)
	err = DBInsert("hash_ca_3", "2026-09-20_10:00:00", "CA", 50.0, 45.0, 2.0, 3.0, 9, 1, 9, 20, "PollsterSep")
	if err != nil {
		t.Fatalf("DBInsert failed: %v", err)
	}

	// TX record in October (different state)
	err = DBInsert("hash_tx_1", "2026-10-12_10:00:00", "TX", 44.0, 52.0, 2.0, 2.0, 10, 5, 10, 12, "PollsterTX")
	if err != nil {
		t.Fatalf("DBInsert failed: %v", err)
	}

	// Test CA report
	outCA := captureOutput(t, func() {
		reportByState("CA")
	})

	// Check header
	if !strings.Contains(outCA, "Start  End    ST  D%   R%   I%   U%  Pollster") {
		t.Fatalf("expected header in output, got:\n%s", outCA)
	}

	// Check Oct records are displayed
	if !strings.Contains(outCA, "10/01  10/15  CA  55   40   03   02  PollsterOct1") {
		t.Fatalf("expected PollsterOct1 row, got:\n%s", outCA)
	}
	if !strings.Contains(outCA, "10/16  10/25  CA  56   39   03   02  PollsterOct2") {
		t.Fatalf("expected PollsterOct2 row, got:\n%s", outCA)
	}

	// Check Sep record is filtered out
	if strings.Contains(outCA, "PollsterSep") {
		t.Fatalf("expected September record to be filtered out, but was present in:\n%s", outCA)
	}

	// Check TX record is not present in CA report
	if strings.Contains(outCA, "PollsterTX") {
		t.Fatalf("expected TX record not to appear in CA report, got:\n%s", outCA)
	}

	// Test TX report
	outTX := captureOutput(t, func() {
		reportByState("TX")
	})

	if !strings.Contains(outTX, "10/05  10/12  TX  44   52   02   02  PollsterTX") {
		t.Fatalf("expected PollsterTX row in TX report, got:\n%s", outTX)
	}
	if strings.Contains(outTX, "PollsterOct1") {
		t.Fatalf("expected CA record not to appear in TX report, got:\n%s", outTX)
	}
}

func TestReportByState_OnlyPreOctoberRecords(t *testing.T) {
	setupTestDB(t)

	// Insert only pre-October record (e.g. August)
	err := DBInsert("hash_ny_1", "2026-08-20_10:00:00", "NY", 60.0, 35.0, 3.0, 2.0, 8, 1, 8, 20, "PollsterAug")
	if err != nil {
		t.Fatalf("DBInsert failed: %v", err)
	}

	out := captureOutput(t, func() {
		reportByState("NY")
	})

	// Header should be printed because records exist, but row should be skipped
	if !strings.Contains(out, "Start  End    ST  D%   R%   I%   U%  Pollster") {
		t.Fatalf("expected header in output, got:\n%s", out)
	}
	if strings.Contains(out, "PollsterAug") {
		t.Fatalf("expected August record to be filtered out, got:\n%s", out)
	}
}
