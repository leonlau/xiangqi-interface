package main

import (
	"strings"
	"testing"

	xq "github.com/leonlau/xiangqi-interface"
)

func TestDecodeMove_UCIValid(t *testing.T) {
	engine := XiangqiEngineImpl{}
	fen := "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"
	m, err := engine.DecodeMove(fen, "h3e3", xq.NotationUCI)
	if err != nil {
		t.Fatalf("DecodeMove(h3e3) error: %v", err)
	}
	if m.S1.File != xq.FileH || m.S1.Rank != xq.Rank2 {
		t.Errorf("S1 = %v, want h2", m.S1)
	}
	if m.S2.File != xq.FileE || m.S2.Rank != xq.Rank2 {
		t.Errorf("S2 = %v, want e2", m.S2)
	}
}

func TestDecodeMove_Chinese(t *testing.T) {
	engine := XiangqiEngineImpl{}
	fen := "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"
	// 炮二平五 = 红炮 H2 → E2
	m, err := engine.DecodeMove(fen, "炮二平五", xq.NotationChinese)
	if err != nil {
		t.Fatalf("DecodeMove(炮二平五) error: %v", err)
	}
	if m.S1 != (xq.Square{File: xq.FileH, Rank: xq.Rank2}) {
		t.Errorf("S1 = %v, want h2", m.S1)
	}
	if m.S2 != (xq.Square{File: xq.FileE, Rank: xq.Rank2}) {
		t.Errorf("S2 = %v, want e2", m.S2)
	}
}

func TestDecodeMove_Invalid(t *testing.T) {
	engine := XiangqiEngineImpl{}
	if _, err := engine.DecodeMove("garbage fen", "h3e3", xq.NotationUCI); err == nil {
		t.Error("invalid FEN should propagate error")
	}
	if _, err := engine.DecodeMove("rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1", "xxxx", xq.NotationUCI); err == nil {
		t.Error("invalid notation should return error")
	}
}

func TestEncodeMove_Format(t *testing.T) {
	engine := XiangqiEngineImpl{}
	fen := "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"
	m := xq.Move{
		S1: xq.Square{File: xq.FileH, Rank: xq.Rank2},
		S2: xq.Square{File: xq.FileE, Rank: xq.Rank2},
	}
	got, err := engine.EncodeMove(fen, m, xq.NotationUCI)
	if err != nil {
		t.Fatalf("EncodeMove error: %v", err)
	}
	if got != "h3e3" {
		t.Errorf("EncodeMove = %q, want %q", got, "h3e3")
	}
}

func TestEncodeMove_Chinese(t *testing.T) {
	engine := XiangqiEngineImpl{}
	fen := "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"
	m := xq.Move{
		S1: xq.Square{File: xq.FileH, Rank: xq.Rank2},
		S2: xq.Square{File: xq.FileE, Rank: xq.Rank2},
	}
	got, err := engine.EncodeMove(fen, m, xq.NotationChinese)
	if err != nil {
		t.Fatalf("EncodeMove(Chinese) error: %v", err)
	}
	if !strings.Contains(got, "炮二平五") {
		t.Errorf("EncodeMove Chinese = %q, want contains 炮二平五", got)
	}
}

func TestEncodeMove_InvalidFEN(t *testing.T) {
	engine := XiangqiEngineImpl{}
	if _, err := engine.EncodeMove("garbage", xq.Move{}, xq.NotationUCI); err == nil {
		t.Error("invalid FEN should return error")
	}
}

func TestDecodeEncodeRoundTrip(t *testing.T) {
	engine := XiangqiEngineImpl{}
	fen := "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"
	for _, uci := range []string{"h3e3", "b1c3", "a1a2"} {
		m, err := engine.DecodeMove(fen, uci, xq.NotationUCI)
		if err != nil {
			if strings.Contains(err.Error(), "chess: move") || strings.Contains(err.Error(), "invalid") {
				continue
			}
			t.Fatalf("DecodeMove(%q): %v", uci, err)
		}
		got, err := engine.EncodeMove(fen, m, xq.NotationUCI)
		if err != nil {
			t.Fatalf("EncodeMove(%v): %v", m, err)
		}
		if got != uci {
			t.Errorf("round-trip %q -> %v -> %q", uci, m, got)
		}
	}
}

func TestEncodeMove_UnknownNotation(t *testing.T) {
	engine := XiangqiEngineImpl{}
	fen := "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"
	if _, err := engine.EncodeMove(fen, xq.Move{}, xq.Notation(99)); err == nil {
		t.Error("unknown notation should return error")
	}
	if _, err := engine.DecodeMove(fen, "h3e3", xq.Notation(99)); err == nil {
		t.Error("unknown notation should return error")
	}
}

func TestUCCINotation(t *testing.T) {
	impl := &XiangqiEngineImpl{}
	startFEN := "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"

	// UCCI 坐标走法 h2e2(红马八进七等)
	mv, err := impl.DecodeMove(startFEN, "h2e2", xq.NotationUCCI)
	if err != nil {
		t.Fatalf("DecodeMove UCCI: %v", err)
	}
	if mv.S1.File != xq.FileH || mv.S1.Rank != xq.Rank2 || mv.S2.File != xq.FileE || mv.S2.Rank != xq.Rank2 {
		t.Fatalf("decoded move: %+v", mv)
	}
	got, err := impl.EncodeMove(startFEN, mv, xq.NotationUCCI)
	if err != nil {
		t.Fatalf("EncodeMove UCCI: %v", err)
	}
	if got != "h2e2" {
		t.Fatalf("encoded UCCI: got %q want %q", got, "h2e2")
	}
}
