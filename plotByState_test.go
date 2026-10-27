package main

import (
	"os"
	"path/filepath"
	"testing"

	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/vgimg"
)

func TestWritePNG(t *testing.T) {
	tempDir := t.TempDir()
	pngPath := filepath.Join(tempDir, "test.png")

	canvas := vgimg.New(2*vg.Inch, 2*vg.Inch)
	err := writePNG(pngPath, canvas)
	if err != nil {
		t.Fatalf("writePNG failed: %v", err)
	}

	info, err := os.Stat(pngPath)
	if err != nil {
		t.Fatalf("expected PNG file to exist at %s: %v", pngPath, err)
	}
	if info.Size() == 0 {
		t.Fatalf("expected PNG file to have non-zero size")
	}
}
