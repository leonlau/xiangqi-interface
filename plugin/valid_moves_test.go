package main

import (
	"testing"

	xq "github.com/leonlau/xiangqi-interface"
)

func TestValidMoves_StartingHasMany(t *testing.T) {
	engine := XiangqiEngineImpl{}
	fen := "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"
	moves, err := engine.ValidMoves(fen)
	if err != nil {
		t.Fatalf("ValidMoves(%q) error: %v", fen, err)
	}
	if len(moves) != 44 {
		t.Errorf("len(moves) = %d, want 44", len(moves))
	}
}

func TestValidMoves_InvalidFEN(t *testing.T) {
	engine := XiangqiEngineImpl{}
	if _, err := engine.ValidMoves("garbage"); err == nil {
		t.Error("ValidMoves(\"garbage\") = nil, want error")
	}
}

func TestValidMoves_AllSquaresInRange(t *testing.T) {
	engine := XiangqiEngineImpl{}
	fen := "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"
	moves, _ := engine.ValidMoves(fen)
	for _, m := range moves {
		if m.S1.File < xq.FileA || m.S1.File > xq.FileI {
			t.Errorf("S1.File out of range: %v", m.S1)
		}
		if m.S2.File < xq.FileA || m.S2.File > xq.FileI {
			t.Errorf("S2.File out of range: %v", m.S2)
		}
		if m.S1.Rank < xq.Rank0 || m.S1.Rank > xq.Rank9 {
			t.Errorf("S1.Rank out of range: %v", m.S1.Rank)
		}
		if m.S2.Rank < xq.Rank0 || m.S2.Rank > xq.Rank9 {
			t.Errorf("S2.Rank out of range: %v", m.S2.Rank)
		}
	}
}
