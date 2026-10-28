package main

import (
	"fmt"
	"testing"
)

func TestSqliteOperations(t *testing.T) {
	setupUnitTest(t)

	// Insert record
	hash := "testhash12345678"
	stamp := "2026-10-26_10:00:00"
	err := DBInsert(hash, stamp, "CA", 55.0, 40.0, 3.0, 2.0, 10, 15, 10, 20, "TestPollster")
	if err != nil {
		t.Fatalf("DBInsert failed: %v", err)
	}

	// Duplicate insert should return an error (due to PRIMARY KEY)
	err = DBInsert(hash, stamp, "CA", 55.0, 40.0, 3.0, 2.0, 10, 15, 10, 20, "TestPollster")
	if err == nil {
		t.Fatalf("expected duplicate insert error, got nil")
	}

	// Query by state
	results, ok := query(fmt.Sprintf(fmtSqlSelectHistoryByState, "CA"))
	if !ok || len(results) != 1 {
		t.Fatalf("expected 1 result, got %d (ok=%v)", len(results), ok)
	}

	// Only 1 result is expected.
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	res := results[0]
	if res.state != "CA" || res.dem != 55.0 || res.gop != 40.0 || res.pollster != "TestPollster" {
		t.Fatalf("unexpected record content: %+v", res)
	}

}
