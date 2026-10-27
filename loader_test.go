package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func setupTestLog(t *testing.T) {
	t.Helper()
	tempDir := t.TempDir()
	openLogFile(tempDir, "test.log")
	t.Cleanup(func() {
		closeLogFile()
	})
}

func TestLoader_Success(t *testing.T) {
	setupTestDB(t)
	setupTestLog(t)

	content := "ak 48 47 00 Oct 14 Oct 15 PPP\n" +
		"tx 45.5 49.5 2.0 Sep 24 Oct 01 Siena U. Polling\n" +
		"fl 50 40 5.0 Jan 01 Jan 02 Mason Dixon\n"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(content))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	startStamp := "2026-10-27_153000"

	loader(startStamp, tempDir, server.URL)

	// Verify downloaded file exists
	downloadedFile := filepath.Join(tempDir, startStamp+".txt")
	fileBytes, err := os.ReadFile(downloadedFile)
	if err != nil {
		t.Fatalf("expected downloaded file %s to exist: %v", downloadedFile, err)
	}
	if string(fileBytes) != content {
		t.Fatalf("downloaded file content mismatch:\nexpected:\n%s\ngot:\n%s", content, string(fileBytes))
	}

	// Verify AK record
	akRecords, ok := query(fmt.Sprintf(fmtSqlSelectHistoryByState, "AK"))
	if !ok || len(akRecords) != 1 {
		t.Fatalf("expected 1 AK record, got %d (ok=%v)", len(akRecords), ok)
	}
	ak := akRecords[0]
	if ak.state != "AK" || ak.dem != 48 || ak.gop != 47 || ak.ind != 0 || ak.unk != 5 ||
		ak.startMonth != 10 || ak.startDay != 14 || ak.endMonth != 10 || ak.endDay != 15 || ak.pollster != "PPP" {
		t.Fatalf("unexpected AK record values: %+v", ak)
	}

	// Verify TX record with float and multi-word pollster
	txRecords, ok := query(fmt.Sprintf(fmtSqlSelectHistoryByState, "TX"))
	if !ok || len(txRecords) != 1 {
		t.Fatalf("expected 1 TX record, got %d (ok=%v)", len(txRecords), ok)
	}
	tx := txRecords[0]
	if tx.state != "TX" || tx.dem != 45.5 || tx.gop != 49.5 || tx.ind != 2.0 || tx.unk != 3.0 ||
		tx.startMonth != 9 || tx.startDay != 24 || tx.endMonth != 10 || tx.endDay != 1 || tx.pollster != "Siena U. Polling" {
		t.Fatalf("unexpected TX record values: %+v", tx)
	}

	// Verify FL record with Jan month
	flRecords, ok := query(fmt.Sprintf(fmtSqlSelectHistoryByState, "FL"))
	if !ok || len(flRecords) != 1 {
		t.Fatalf("expected 1 FL record, got %d (ok=%v)", len(flRecords), ok)
	}
	fl := flRecords[0]
	if fl.state != "FL" || fl.dem != 50 || fl.gop != 40 || fl.ind != 5.0 || fl.unk != 5.0 ||
		fl.startMonth != 1 || fl.startDay != 1 || fl.endMonth != 1 || fl.endDay != 2 || fl.pollster != "Mason Dixon" {
		t.Fatalf("unexpected FL record values: %+v", fl)
	}
}

func TestLoader_DuplicatesAndDiscards(t *testing.T) {
	setupTestDB(t)
	setupTestLog(t)

	content1 := "CA 55 40 00 Oct 01 Oct 05 PollsterOne\n" +
		"NY 60 30 00 Oct 02 Oct 06 PollsterTwo\n"

	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(content1))
	}))
	defer server1.Close()

	tempDir := t.TempDir()
	loader("2026-10-27_100000", tempDir, server1.URL)

	// Load second time with 1 existing record and 1 new record
	content2 := "CA 55 40 00 Oct 01 Oct 05 PollsterOne\n" + // duplicate hash
		"NV 48 48 00 Oct 03 Oct 07 PollsterThree\n" // new

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(content2))
	}))
	defer server2.Close()

	loader("2026-10-27_110000", tempDir, server2.URL)

	// Verify total count for CA is still 1
	caRecords, ok := query(fmt.Sprintf(fmtSqlSelectHistoryByState, "CA"))
	if !ok || len(caRecords) != 1 {
		t.Fatalf("expected 1 CA record, got %d (ok=%v)", len(caRecords), ok)
	}

	// Verify NY and NV records
	nyRecords, ok := query(fmt.Sprintf(fmtSqlSelectHistoryByState, "NY"))
	if !ok || len(nyRecords) != 1 {
		t.Fatalf("expected 1 NY record, got %d (ok=%v)", len(nyRecords), ok)
	}
	nvRecords, ok := query(fmt.Sprintf(fmtSqlSelectHistoryByState, "NV"))
	if !ok || len(nvRecords) != 1 {
		t.Fatalf("expected 1 NV record, got %d (ok=%v)", len(nvRecords), ok)
	}
}

func TestLoader_WindowsLineEndings(t *testing.T) {
	setupTestDB(t)
	setupTestLog(t)

	// Test CRLF handling
	content := "OH 46 48 02 Jul 10 Jul 15 PollsterCRLF\r\nMI 50 45 00 Aug 01 Aug 05 PollsterMI\r\n"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(content))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	loader("2026-10-27_120000", tempDir, server.URL)

	ohRecords, ok := query(fmt.Sprintf(fmtSqlSelectHistoryByState, "OH"))
	if !ok || len(ohRecords) != 1 {
		t.Fatalf("expected 1 OH record, got %d (ok=%v)", len(ohRecords), ok)
	}
	if ohRecords[0].pollster != "PollsterCRLF" {
		t.Fatalf("unexpected pollster name: %q", ohRecords[0].pollster)
	}

	miRecords, ok := query(fmt.Sprintf(fmtSqlSelectHistoryByState, "MI"))
	if !ok || len(miRecords) != 1 {
		t.Fatalf("expected 1 MI record, got %d (ok=%v)", len(miRecords), ok)
	}
	if miRecords[0].pollster != "PollsterMI" {
		t.Fatalf("unexpected pollster name: %q", miRecords[0].pollster)
	}
}
