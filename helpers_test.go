package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestGetUtcDate(t *testing.T) {
	d := getUtcDate()
	if len(d) != 10 || d[4] != '-' || d[7] != '-' {
		t.Fatalf("unexpected UTC date format: %q", d)
	}
}

func TestGetUtcTime(t *testing.T) {
	tm := getUtcTime()
	if len(tm) != 12 || tm[2] != ':' || tm[5] != ':' || tm[8] != '.' {
		t.Fatalf("unexpected UTC time format: %q", tm)
	}
}

func TestCleanerText(t *testing.T) {
	// Without newlines preserved
	input := "Hello\x00\x07World\n\r"
	cleaned := cleanerText(input, false)
	if cleaned != "HelloWorld" {
		t.Fatalf("expected HelloWorld, got %q", cleaned)
	}

	// With newlines preserved
	cleanedWithNL := cleanerText(input, true)
	if cleanedWithNL != "Hello  World\n\r" {
		t.Fatalf("expected 'Hello  World\\n\\r', got %q", cleanedWithNL)
	}
}

func TestSortCaseInsensitive(t *testing.T) {
	items := []string{"banana", "Apple", "cherry", "date"}
	sortCaseInsensitive(&items)
	expected := []string{"Apple", "banana", "cherry", "date"}
	if !slices.Equal(items, expected) {
		t.Fatalf("expected %v, got %v", expected, items)
	}
}

func TestHashStrArray(t *testing.T) {
	arr1 := []string{"field1", "field2"}
	h1 := hashStrArray(arr1)
	if len(h1) != 32 {
		t.Fatalf("expected 32 hex chars, got %d (%q)", len(h1), h1)
	}

	arr2 := []string{"field1", "field2"}
	h2 := hashStrArray(arr2)
	if h1 != h2 {
		t.Fatalf("expected hashes to match, got %q vs %q", h1, h2)
	}

	arr3 := []string{"field1", "field3"}
	h3 := hashStrArray(arr3)
	if h1 == h3 {
		t.Fatalf("expected hashes to differ, got identical %q", h1)
	}
}

func TestComputeUnknown(t *testing.T) {
	unk := computeUnknown(45.0, 40.0, 5.0)
	if unk != 10.0 {
		t.Fatalf("expected 10.0, got %f", unk)
	}

	unk2 := computeUnknown(50.0, 50.0, 0.0)
	if unk2 != 0.0 {
		t.Fatalf("expected 0.0, got %f", unk2)
	}
}

func TestStringToFloat64(t *testing.T) {
	val := stringToFloat64("45.67", "dem")
	if val != 45.67 {
		t.Fatalf("expected 45.67, got %f", val)
	}
}

func TestStringToInt64(t *testing.T) {
	val := stringToInt64("12345", "startDay")
	if val != 12345 {
		t.Fatalf("expected 12345, got %d", val)
	}
}

func TestMonthStringToInt64(t *testing.T) {
	months := map[string]int64{
		"Jan": 1, "Feb": 2, "Mar": 3, "Apr": 4,
		"May": 5, "Jun": 6, "Jul": 7, "Aug": 8,
		"Sep": 9, "Oct": 10, "Nov": 11, "Dec": 12,
	}

	for mStr, mVal := range months {
		got := monthStringToInt64(mStr, "month")
		if got != mVal {
			t.Fatalf("for month %q expected %d, got %d", mStr, mVal, got)
		}
	}
}

func TestIsDirectory(t *testing.T) {
	tempDir := t.TempDir()
	if err := isDirectory(tempDir); err != nil {
		t.Fatalf("expected tempDir to be a directory: %v", err)
	}

	filePath := filepath.Join(tempDir, "sample.txt")
	if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	if err := isDirectory(filePath); err == nil {
		t.Fatalf("expected error for file, got nil")
	}

	toBeCreated := filepath.Join(tempDir, "to-be-created")
	if err := isDirectory(toBeCreated); err != nil {
		t.Fatalf("expected non-existent dir path to be created: %v", err)
	}
}

func TestMakeDirAndStoreTextAndWriteOutputText(t *testing.T) {
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "subdir")
	makeDir(subDir)

	if err := isDirectory(subDir); err != nil {
		t.Fatalf("makeDir failed to create directory: %v", err)
	}

	// Idempotent makeDir
	makeDir(subDir)

	// storeText
	storeText(subDir, "test.txt", "hello world")
	content, err := os.ReadFile(filepath.Join(subDir, "test.txt"))
	if err != nil {
		t.Fatalf("storeText failed: %v", err)
	}
	if !strings.Contains(string(content), "hello world") {
		t.Fatalf("unexpected content: %q", string(content))
	}

	// writeOutputText
	f, err := os.Create(filepath.Join(subDir, "out.txt"))
	if err != nil {
		t.Fatalf("os.Create failed: %v", err)
	}
	writeOutputText(f, "line 1")
	_ = f.Close()

	outContent, err := os.ReadFile(filepath.Join(subDir, "out.txt"))
	if err != nil {
		t.Fatalf("failed to read out.txt: %v", err)
	}
	if !strings.Contains(string(outContent), "line 1") {
		t.Fatalf("unexpected content in out.txt: %q", string(outContent))
	}
}

func TestUrlGet(t *testing.T) {
	content := "test content for urlGet\nline 2"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(content))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	outPath := filepath.Join(tempDir, "downloaded.txt")

	urlGet(server.URL, outPath)

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}
	if string(data) != content {
		t.Fatalf("expected content %q, got %q", content, string(data))
	}
}
