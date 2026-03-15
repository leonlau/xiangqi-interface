package main

import (
	"testing"

	"github.com/leonlau/xiangqi"
	xq "github.com/leonlau/xiangqi-interface"
)

func TestParsePosition_StartingPosition(t *testing.T) {
	fen := "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"
	pos, err := parsePosition(fen)
	if err != nil {
		t.Fatalf("parsePosition(%q) error: %v", fen, err)
	}
	if pos == nil {
		t.Fatal("parsePosition returned nil position")
	}
}

func TestConvertMove_RoundTrip(t *testing.T) {
	pos := xiangqi.StartingPosition()
	moves := pos.ValidMoves()
	if len(moves) == 0 {
		t.Fatal("starting position has no valid moves")
	}
	for _, m := range moves {
		got := convertMove(m)
		// S1 / S2 应该是有效的 a0..i9
		if got.S1.File < xq.FileA || got.S1.File > xq.FileI {
			t.Errorf("S1.File out of range: %v", got.S1.File)
		}
		if got.S2.File < xq.FileA || got.S2.File > xq.FileI {
			t.Errorf("S2.File out of range: %v", got.S2.File)
		}
	}
}

func TestFormatUCI_RankOffset(t *testing.T) {
	cases := []struct {
		in   xq.Square
		want string
	}{
		{xq.Square{File: xq.FileA, Rank: xq.Rank0}, "a1"},
		{xq.Square{File: xq.FileH, Rank: xq.Rank2}, "h3"},
		{xq.Square{File: xq.FileI, Rank: xq.Rank9}, "i10"},
	}
	for _, c := range cases {
		got := formatUCI(c.in)
		if got != c.want {
			t.Errorf("formatUCI(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}
