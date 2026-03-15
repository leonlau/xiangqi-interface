package main

import (
	"os"
	"strings"
	"testing"

	xq "github.com/leonlau/xiangqi-interface"
)

func TestParsePGN_Basic(t *testing.T) {
	data, err := os.ReadFile("../testdata/sample.pgn")
	if err != nil {
		t.Fatalf("read sample.pgn: %v", err)
	}
	engine := XiangqiEngineImpl{}
	game, err := engine.ParsePGN(string(data))
	if err != nil {
		t.Fatalf("ParsePGN: %v", err)
	}

	if len(game.Moves) != 1 {
		t.Errorf("len(Moves) = %d, want 1", len(game.Moves))
	}
	if game.Result != "*" {
		t.Errorf("Result = %q, want *", game.Result)
	}

	gotEvent := ""
	for _, tag := range game.Tags {
		if tag.Name == "Event" {
			gotEvent = tag.Value
		}
	}
	if gotEvent != "Test" {
		t.Errorf("Event tag = %q, want Test", gotEvent)
	}

	if len(game.Moves) > 0 {
		first := game.Moves[0]
		if first.S1.File != xq.FileH || first.S1.Rank != xq.Rank2 {
			t.Errorf("Moves[0].S1 = %v, want h2", first.S1)
		}
		if first.S2.File != xq.FileE || first.S2.Rank != xq.Rank2 {
			t.Errorf("Moves[0].S2 = %v, want e2", first.S2)
		}
	}
}

func TestParsePGN_Invalid(t *testing.T) {
	engine := XiangqiEngineImpl{}
	if _, err := engine.ParsePGN("this is not pgn at all {{}{}{"); err == nil {
		t.Error("invalid PGN should return error")
	}
}

func TestEncodePGN_Empty(t *testing.T) {
	engine := XiangqiEngineImpl{}
	out, err := engine.EncodePGN(xq.PGNGame{
		Tags:   []xq.PGNTag{{Name: "Event", Value: "Test"}},
		Moves:  nil,
		Result: "*",
	})
	if err != nil {
		t.Fatalf("EncodePGN: %v", err)
	}
	if !strings.Contains(out, `[Event "Test"]`) {
		t.Errorf("output missing Event tag: %q", out)
	}
}

func TestEncodePGN_WithMoves(t *testing.T) {
	engine := XiangqiEngineImpl{}
	game := xq.PGNGame{
		Tags: []xq.PGNTag{
			{Name: "Event", Value: "Test"},
			{Name: "Result", Value: "*"},
		},
		Moves: []xq.Move{
			{S1: xq.Square{File: xq.FileH, Rank: xq.Rank2}, S2: xq.Square{File: xq.FileE, Rank: xq.Rank2}},
		},
		Result: "*",
	}
	out, err := engine.EncodePGN(game)
	if err != nil {
		t.Fatalf("EncodePGN: %v", err)
	}
	if !strings.Contains(out, "炮二平五") {
		t.Errorf("output missing move 炮二平五: %q", out)
	}
}
