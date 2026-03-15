//go:build integration

package sdk

import (
	"os"
	"testing"

	xq "github.com/leonlau/xiangqi-interface"
)

const startFEN = "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"

func mustLoadSDK(t *testing.T) *SDK {
	t.Helper()
	path := os.Getenv("ENGINE_PLUGIN")
	if path == "" {
		path = "../dist/libengine.so"
	}
	sdk, err := New(path)
	if err != nil {
		t.Skipf("plugin not built: %v", err)
	}
	return sdk
}

func TestIntegration_LoadAndCallValidMoves(t *testing.T) {
	sdk := mustLoadSDK(t)
	moves, err := sdk.ValidMoves(startFEN)
	if err != nil {
		t.Fatalf("ValidMoves: %v", err)
	}
	if len(moves) != 44 {
		t.Errorf("len(moves) = %d, want 44", len(moves))
	}
}

func TestIntegration_DecodeEncodeUCIRoundTrip(t *testing.T) {
	sdk := mustLoadSDK(t)
	m, err := sdk.DecodeMove(startFEN, "h3e3", xq.NotationUCI)
	if err != nil {
		t.Fatalf("DecodeMove: %v", err)
	}
	got, err := sdk.EncodeMove(startFEN, m, xq.NotationUCI)
	if err != nil {
		t.Fatalf("EncodeMove: %v", err)
	}
	if got != "h3e3" {
		t.Errorf("round-trip = %q, want h3e3", got)
	}
}

func TestIntegration_ChineseNotation(t *testing.T) {
	sdk := mustLoadSDK(t)
	m, err := sdk.DecodeMove(startFEN, "炮二平五", xq.NotationChinese)
	if err != nil {
		t.Fatalf("DecodeMove(炮二平五): %v", err)
	}
	encoded, err := sdk.EncodeMove(startFEN, m, xq.NotationChinese)
	if err != nil {
		t.Fatalf("EncodeMove(Chinese): %v", err)
	}
	if encoded == "" {
		t.Error("EncodeMove(Chinese) returned empty")
	}
}

func TestIntegration_ApplyMove(t *testing.T) {
	sdk := mustLoadSDK(t)
	m := xq.Move{S1: xq.Square{File: xq.FileH, Rank: xq.Rank2}, S2: xq.Square{File: xq.FileE, Rank: xq.Rank2}}
	newFen, err := sdk.ApplyMove(startFEN, m)
	if err != nil {
		t.Fatalf("ApplyMove: %v", err)
	}
	turn, _ := sdk.Turn(newFen)
	if turn != xq.Black {
		t.Errorf("after ApplyMove, turn = %v, want Black", turn)
	}
}

func TestIntegration_BoardAndHash(t *testing.T) {
	sdk := mustLoadSDK(t)
	board, err := sdk.Board(startFEN)
	if err != nil {
		t.Fatalf("Board: %v", err)
	}
	occupied := 0
	for _, p := range board {
		if p.Type != xq.NoPieceType {
			occupied++
		}
	}
	if occupied != 32 {
		t.Errorf("occupied = %d, want 32", occupied)
	}
	h, err := sdk.Hash(startFEN)
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	if len(h) != 32 {
		t.Errorf("hash length = %d, want 32", len(h))
	}
}

func TestIntegration_GameLifecycle(t *testing.T) {
	sdk := mustLoadSDK(t)
	h, err := sdk.NewGame(startFEN)
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	defer sdk.CloseGame(h)

	m := xq.Move{S1: xq.Square{File: xq.FileH, Rank: xq.Rank2}, S2: xq.Square{File: xq.FileE, Rank: xq.Rank2}}
	if err := sdk.GameMove(h, m); err != nil {
		t.Fatalf("GameMove: %v", err)
	}
	fen, _ := sdk.GameFEN(h)
	if fen == "" {
		t.Error("GameFEN returned empty")
	}
	moves, _ := sdk.GameMoves(h)
	if len(moves) != 1 {
		t.Errorf("GameMoves = %d, want 1", len(moves))
	}
}

func TestIntegration_OpeningFind(t *testing.T) {
	sdk := mustLoadSDK(t)
	moves := []xq.Move{
		{S1: xq.Square{File: xq.FileH, Rank: xq.Rank2}, S2: xq.Square{File: xq.FileE, Rank: xq.Rank2}},
	}
	o, err := sdk.OpeningFind(moves)
	if err != nil {
		t.Fatalf("OpeningFind: %v", err)
	}
	if o == nil {
		t.Skip("no opening found in book (may be expected for some positions)")
	}
	t.Logf("found opening: %s / %s", o.Code, o.Name)
}

func TestIntegration_GameExtended(t *testing.T) {
	sdk := mustLoadSDK(t)
	h, _ := sdk.NewGame(startFEN)
	defer sdk.CloseGame(h)

	// Tag CRUD: xiangqi AddTagPair 返回 false=新加,true=已存在更新
	if _, err := sdk.GameTagAdd(h, "Site", "test"); err != nil {
		t.Fatalf("GameTagAdd: %v", err)
	}
	if v, found, err := sdk.GameTagGet(h, "Site"); err != nil || !found || v != "test" {
		t.Errorf("GameTagGet: v=%q found=%v err=%v", v, found, err)
	}
	sdk.GameTagRemove(h, "Site")
	if _, found, _ := sdk.GameTagGet(h, "Site"); found {
		t.Error("Site tag still present after remove")
	}

	// GamePositions
	m := xq.Move{S1: xq.Square{File: xq.FileH, Rank: xq.Rank2}, S2: xq.Square{File: xq.FileE, Rank: xq.Rank2}}
	sdk.GameMove(h, m)
	positions, err := sdk.GamePositions(h)
	if err != nil {
		t.Fatalf("GamePositions: %v", err)
	}
	if len(positions) != 2 {
		t.Errorf("GamePositions count = %d, want 2 (start + after move)", len(positions))
	}

	// GameClone
	h2, err := sdk.GameClone(h)
	if err != nil {
		t.Fatalf("GameClone: %v", err)
	}
	defer sdk.CloseGame(h2)
	if h == h2 {
		t.Error("cloned handle equals original")
	}
}

func TestIntegration_BoardFlip(t *testing.T) {
	sdk := mustLoadSDK(t)
	updown, err := sdk.BoardFlip(startFEN, xq.FlipUpDown)
	if err != nil {
		t.Fatalf("BoardFlip UpDown: %v", err)
	}
	// rank 0 of original = RNBAKABNR (red back rank)
	// After UpDown, rank 9 of new = RNBAKABNR (moved to top)
	if p := updown[9*9+0]; p.Type != xq.Rook || p.Color != xq.Red {
		t.Errorf("After UpDown, board[9][0] = %v, want Red Rook", p)
	}
}

func TestIntegration_PositionMarshal(t *testing.T) {
	sdk := mustLoadSDK(t)
	text, err := sdk.PositionMarshalText(startFEN)
	if err != nil {
		t.Fatalf("PositionMarshalText: %v", err)
	}
	roundtrip, err := sdk.PositionUnmarshalText(text)
	if err != nil {
		t.Fatalf("PositionUnmarshalText: %v", err)
	}
	if roundtrip != startFEN {
		t.Errorf("round-trip = %q, want %q", roundtrip, startFEN)
	}
}

func TestIntegration_GameJudge(t *testing.T) {
	sdk := mustLoadSDK(t)
	h, _ := sdk.NewGame(startFEN)
	defer sdk.CloseGame(h)
	v, err := sdk.GameJudge(h)
	if err != nil {
		t.Fatalf("GameJudge: %v", err)
	}
	// 起始局面无裁决
	if v.Outcome != xq.OutcomeUnknown {
		t.Errorf("starting position verdict outcome = %v, want Unknown", v.Outcome)
	}
	t.Logf("starting verdict: outcome=%v method=%v reason=%q", v.Outcome, v.Method, v.Reason)
}

func TestIntegration_GameMoveHistory(t *testing.T) {
	sdk := mustLoadSDK(t)
	h, _ := sdk.NewGame(startFEN)
	defer sdk.CloseGame(h)
	m := xq.Move{S1: xq.Square{File: xq.FileH, Rank: xq.Rank2}, S2: xq.Square{File: xq.FileE, Rank: xq.Rank2}}
	sdk.GameMove(h, m)
	history, err := sdk.GameMoveHistory(h)
	if err != nil {
		t.Fatalf("GameMoveHistory: %v", err)
	}
	if len(history) != 1 {
		t.Errorf("history len = %d, want 1", len(history))
	}
	h0 := history[0]
	if h0.PreFEN != startFEN {
		t.Errorf("PreFEN = %q, want %q", h0.PreFEN, startFEN)
	}
	if h0.Move != m {
		t.Errorf("Move = %+v, want %+v", h0.Move, m)
	}
	if h0.PostFEN == "" {
		t.Error("PostFEN is empty")
	}
}

func TestIntegration_UCIEngine(t *testing.T) {
	sdk := mustLoadSDK(t)
	// 找 fairy-stockfish 可执行
	enginePath := os.Getenv("ENGINE_UCI_BIN")
	if enginePath == "" {
		t.Skip("ENGINE_UCI_BIN not set; skipping UCI integration test")
	}
	h, err := sdk.NewEngine(enginePath)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	defer sdk.EngineClose(h)

	// 正确握手: uci → isready → ucinewgame
	if err := sdk.EngineRun(h, xq.UCICmdUCI{}); err != nil {
		t.Fatalf("EngineRun uci: %v", err)
	}
	if err := sdk.EngineRun(h, xq.UCICmdIsReady{}); err != nil {
		t.Fatalf("EngineRun isready: %v", err)
	}

	id, err := sdk.EngineID(h)
	if err != nil {
		t.Fatalf("EngineID: %v", err)
	}
	if _, ok := id["name"]; !ok {
		t.Errorf("EngineID missing 'name': %v", id)
	}
	t.Logf("engine: name=%s author=%s", id["name"], id["author"])

	opts, err := sdk.EngineOptions(h)
	if err != nil {
		t.Fatalf("EngineOptions: %v", err)
	}
	t.Logf("engine has %d options", len(opts))

	// 走 1 步后搜 depth 1
	if err := sdk.EngineRun(h, xq.UCICmdPosition{StartFEN: startFEN}); err != nil {
		t.Fatalf("EngineRun position: %v", err)
	}
	if err := sdk.EngineRun(h, xq.UCICmdGo{Depth: 1}); err != nil {
		t.Fatalf("EngineRun go: %v", err)
	}
	res, err := sdk.EngineSearchResults(h)
	if err != nil {
		t.Fatalf("EngineSearchResults: %v", err)
	}
	if res.BestMove == "" {
		t.Error("BestMove is empty")
	}
	t.Logf("search result: bestmove=%s info=%+v", res.BestMove, res.Info)
}
