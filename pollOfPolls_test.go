package main

import (
	"strings"
	"testing"
)

func TestAppendWinner(t *testing.T) {

	setupUnitTest(t)

	// Zero count should be ignored
	w := &typeWinner{}
	appendWinner(w, "CA", typeStateTotals{count: 0})
	if w.demCount != 0 || w.gopCount != 0 || w.indCount != 0 || w.unkCount != 0 {
		t.Fatalf("expected all counts to remain 0, got %+v", w)
	}

	// Dem win
	appendWinner(w, "CA", typeStateTotals{count: 1, dem: 60, gop: 30, ind: 5, unk: 5})
	if w.demCount != 1 || !strings.Contains(w.demString, "CA") {
		t.Fatalf("expected dem win for CA, got %+v", w)
	}

	// GOP win
	appendWinner(w, "TX", typeStateTotals{count: 1, dem: 40, gop: 55, ind: 2, unk: 3})
	if w.gopCount != 1 || !strings.Contains(w.gopString, "TX") {
		t.Fatalf("expected gop win for TX, got %+v", w)
	}

	// Ind win
	appendWinner(w, "ME", typeStateTotals{count: 1, dem: 20, gop: 25, ind: 50, unk: 5})
	if w.indCount != 1 || !strings.Contains(w.indString, "ME") {
		t.Fatalf("expected ind win for ME, got %+v", w)
	}

	// Unk win / tie
	appendWinner(w, "OH", typeStateTotals{count: 1, dem: 25, gop: 25, ind: 10, unk: 40})
	if w.unkCount != 1 || !strings.Contains(w.unkString, "OH") {
		t.Fatalf("expected unk win for OH, got %+v", w)
	}
}

func TestPollStateResultEmit(t *testing.T) {

	setupUnitTest(t)

	// pollStateResultEmit should handle count < 1.0 gracefully without panic
	pollStateResultEmit("CA", &typeStateTotals{count: 0}, 2.0)

	// pollStateResultEmit when unk < gopPctIncrement
	pollStateResultEmit("TX", &typeStateTotals{count: 2, dem: 80, gop: 100, ind: 10, unk: 10}, 10.0)

	// pollStateResultEmit when unk >= gopPctIncrement
	pollStateResultEmit("FL", &typeStateTotals{count: 1, dem: 45, gop: 45, ind: 0, unk: 10}, 3.0)
}
