package main

import (
	"strings"
	"testing"
)

func TestReportByState_NoRecords(t *testing.T) {
	setupUnitTest(t)

	out := captureOutput(t, func() {
		reportByState("NV")
	})

	expected := "No history records were found for state NV"
	if !strings.Contains(out, expected) {
		t.Fatalf("expected output to contain %q, got:\n%s", expected, out)
	}
}

func TestReportByState_WithRecords(t *testing.T) {
	setupUnitTest(t)

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
	setupUnitTest(t)

	// Insert only pre-October record (e.g. August)
	err := DBInsert("hash_ny_1", "2026-10-20_10:00:00", "NY", 60.0, 35.0, 3.0, 2.0, 8, 1, 8, 20, "PollsterOct")
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
	if !strings.Contains(out, "PollsterOct") {
		t.Fatalf("Missing PollsterOct record, saw:\n%s", out)
	}
}
