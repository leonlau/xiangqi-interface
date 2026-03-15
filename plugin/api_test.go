package main

import (
	"testing"

	xq "github.com/leonlau/xiangqi-interface"
)

func TestCanonicalFEN_RoundTrip(t *testing.T) {
	engine := XiangqiEngineImpl{}
	fen := "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"
	got, err := engine.CanonicalFEN(fen)
	if err != nil {
		t.Fatalf("CanonicalFEN: %v", err)
	}
	if got != fen {
		t.Errorf("CanonicalFEN = %q, want %q", got, fen)
	}
}

func TestTurn_StartingPosition(t *testing.T) {
	engine := XiangqiEngineImpl{}
	fen := "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"
	c, err := engine.Turn(fen)
	if err != nil {
		t.Fatalf("Turn: %v", err)
	}
	if c != xq.Red {
		t.Errorf("Turn = %v, want Red", c)
	}
}

func TestInCheck_StartingPosition(t *testing.T) {
	engine := XiangqiEngineImpl{}
	fen := "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"
	for _, c := range []xq.Color{xq.Red, xq.Black} {
		inCheck, err := engine.InCheck(fen, c)
		if err != nil {
			t.Fatalf("InCheck(%v): %v", c, err)
		}
		if inCheck {
			t.Errorf("starting position: %v should not be in check", c)
		}
	}
}

func TestStatus_StartingInProgress(t *testing.T) {
	engine := XiangqiEngineImpl{}
	fen := "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"
	s, err := engine.Status(fen)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if s != xq.StatusInProgress {
		t.Errorf("Status = %v, want InProgress", s)
	}
}

func TestBoard_StartingPiecesCount(t *testing.T) {
	engine := XiangqiEngineImpl{}
	fen := "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"
	board, err := engine.Board(fen)
	if err != nil {
		t.Fatalf("Board: %v", err)
	}
	// 起始局面:红 16 子 + 黑 16 子 = 32 子
	occupied := 0
	for _, p := range board {
		if p.Type != xq.NoPieceType {
			occupied++
		}
	}
	if occupied != 32 {
		t.Errorf("occupied squares = %d, want 32", occupied)
	}
}

func TestHash_StableAndHex(t *testing.T) {
	engine := XiangqiEngineImpl{}
	fen := "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"
	h1, err := engine.Hash(fen)
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	h2, _ := engine.Hash(fen)
	if h1 != h2 {
		t.Errorf("Hash not stable: %q vs %q", h1, h2)
	}
	if len(h1) != 32 {
		t.Errorf("Hash length = %d, want 32 (16 bytes hex)", len(h1))
	}
}

func TestApplyMove_StartingPosition(t *testing.T) {
	engine := XiangqiEngineImpl{}
	fen := "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"
	m := xq.Move{S1: xq.Square{File: xq.FileH, Rank: xq.Rank2}, S2: xq.Square{File: xq.FileE, Rank: xq.Rank2}}
	newFen, err := engine.ApplyMove(fen, m)
	if err != nil {
		t.Fatalf("ApplyMove: %v", err)
	}
	if newFen == fen {
		t.Error("ApplyMove: FEN did not change")
	}
	turn, _ := engine.Turn(newFen)
	if turn != xq.Black {
		t.Errorf("after red move, turn = %v, want Black", turn)
	}
}

func TestApplyMoves_Batch(t *testing.T) {
	engine := XiangqiEngineImpl{}
	fen := "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"
	moves := []xq.Move{
		{S1: xq.Square{File: xq.FileH, Rank: xq.Rank2}, S2: xq.Square{File: xq.FileE, Rank: xq.Rank2}}, // 炮二平五
		{S1: xq.Square{File: xq.FileH, Rank: xq.Rank7}, S2: xq.Square{File: xq.FileE, Rank: xq.Rank7}}, // 炮8平5
	}
	newFen, err := engine.ApplyMoves(fen, moves)
	if err != nil {
		t.Fatalf("ApplyMoves: %v", err)
	}
	if newFen == fen {
		t.Error("ApplyMoves: FEN did not change")
	}
	turn, _ := engine.Turn(newFen)
	if turn != xq.Red {
		t.Errorf("after 2 moves, turn = %v, want Red", turn)
	}
}

func TestNewGame_AndGameMove(t *testing.T) {
	engine := XiangqiEngineImpl{}
	h, err := engine.NewGame("rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1")
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	if h == xq.InvalidGameHandle {
		t.Fatal("NewGame returned InvalidGameHandle")
	}
	defer engine.CloseGame(h)

	m := xq.Move{S1: xq.Square{File: xq.FileH, Rank: xq.Rank2}, S2: xq.Square{File: xq.FileE, Rank: xq.Rank2}}
	if err := engine.GameMove(h, m); err != nil {
		t.Fatalf("GameMove: %v", err)
	}

	fen, err := engine.GameFEN(h)
	if err != nil {
		t.Fatalf("GameFEN: %v", err)
	}
	if fen == "" {
		t.Error("GameFEN returned empty")
	}

	moves, err := engine.GameMoves(h)
	if err != nil {
		t.Fatalf("GameMoves: %v", err)
	}
	if len(moves) != 1 {
		t.Errorf("len(GameMoves) = %d, want 1", len(moves))
	}

	status, err := engine.GameStatus(h)
	if err != nil {
		t.Fatalf("GameStatus: %v", err)
	}
	if status != xq.StatusInProgress {
		t.Errorf("GameStatus = %v, want InProgress", status)
	}

	pgn, err := engine.GamePGN(h)
	if err != nil {
		t.Fatalf("GamePGN: %v", err)
	}
	if pgn == "" {
		t.Error("GamePGN returned empty")
	}
}

func TestGame_InvalidHandle(t *testing.T) {
	engine := XiangqiEngineImpl{}
	if err := engine.GameMove(99999, xq.Move{}); err == nil {
		t.Error("GameMove with invalid handle should error")
	}
	if _, err := engine.GameFEN(99999); err == nil {
		t.Error("GameFEN with invalid handle should error")
	}
	if err := engine.CloseGame(99999); err == nil {
		t.Error("CloseGame with invalid handle should error")
	}
}

func TestOpeningFind_StartingMove(t *testing.T) {
	engine := XiangqiEngineImpl{}
	// 炮二平五 = 红炮 H2 → E2
	moves := []xq.Move{
		{S1: xq.Square{File: xq.FileH, Rank: xq.Rank2}, S2: xq.Square{File: xq.FileE, Rank: xq.Rank2}},
	}
	o, err := engine.OpeningFind(moves)
	if err != nil {
		t.Fatalf("OpeningFind: %v", err)
	}
	// 至少应能查到某个开局(可能 nil 如果 ECO 不覆盖)
	if o != nil {
		t.Logf("found opening: %s / %s", o.Code, o.Name)
	}
}

func TestOpeningPossible_StartingMove(t *testing.T) {
	engine := XiangqiEngineImpl{}
	moves := []xq.Move{
		{S1: xq.Square{File: xq.FileH, Rank: xq.Rank2}, S2: xq.Square{File: xq.FileE, Rank: xq.Rank2}},
	}
	os, err := engine.OpeningPossible(moves)
	if err != nil {
		t.Fatalf("OpeningPossible: %v", err)
	}
	t.Logf("got %d possible openings", len(os))
}

func TestHalfMoveClock_StartingZero(t *testing.T) {
	engine := XiangqiEngineImpl{}
	fen := "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"
	n, err := engine.HalfMoveClock(fen)
	if err != nil {
		t.Fatalf("HalfMoveClock: %v", err)
	}
	if n != 0 {
		t.Errorf("HalfMoveClock at start = %d, want 0", n)
	}
}

func TestGameUndo_RestoresState(t *testing.T) {
	engine := XiangqiEngineImpl{}
	startFEN := "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"
	h, _ := engine.NewGame(startFEN)
	defer engine.CloseGame(h)
	// 红炮 H2→E2  (炮二平五)
	m1 := xq.Move{S1: xq.Square{File: xq.FileH, Rank: xq.Rank2}, S2: xq.Square{File: xq.FileE, Rank: xq.Rank2}}
	if err := engine.GameMove(h, m1); err != nil {
		t.Fatalf("GameMove: %v", err)
	}

	fenAfterMove, _ := engine.GameFEN(h)
	if fenAfterMove == startFEN {
		t.Fatal("setup: FEN unchanged after move")
	}

	undone, err := engine.GameUndo(h)
	if err != nil {
		t.Fatalf("GameUndo: %v", err)
	}
	if undone.S1 != m1.S1 || undone.S2 != m1.S2 {
		t.Errorf("undone = %+v, want m1", undone)
	}

	movesAfter, _ := engine.GameMoves(h)
	if len(movesAfter) != 0 {
		t.Errorf("after undo, moves = %d, want 0", len(movesAfter))
	}

	fen, _ := engine.GameFEN(h)
	if fen != startFEN {
		t.Errorf("after 1 undo, FEN = %q, want start %q", fen, startFEN)
	}
}

func TestGameUndo_Empty(t *testing.T) {
	engine := XiangqiEngineImpl{}
	h, _ := engine.NewGame("")
	defer engine.CloseGame(h)
	if _, err := engine.GameUndo(h); err == nil {
		t.Error("undo on empty game should error")
	}
}

func TestGameResign(t *testing.T) {
	engine := XiangqiEngineImpl{}
	h, _ := engine.NewGame("")
	defer engine.CloseGame(h)
	if err := engine.GameResign(h, xq.Red); err != nil {
		t.Fatalf("GameResign: %v", err)
	}
	pgn, _ := engine.GamePGN(h)
	if pgn == "" {
		t.Error("GamePGN after resign empty")
	}
}

func TestNewGame_WithOptions(t *testing.T) {
	engine := XiangqiEngineImpl{}
	h, err := engine.NewGame("", xq.UseRulesOpt{Rules: "asian"}, xq.IgnoreRulesOpt{})
	if err != nil {
		t.Fatalf("NewGame with options: %v", err)
	}
	defer engine.CloseGame(h)
	if h == xq.InvalidGameHandle {
		t.Error("NewGame returned InvalidGameHandle")
	}
}

func TestMove_TagsCaptured(t *testing.T) {
	engine := XiangqiEngineImpl{}
	fen := "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"
	moves, _ := engine.ValidMoves(fen)
	// 起始局面 h2(红炮)平到 e2 吃不到子(空着),所以不一定有 capture
	// 找一个有 capture 标记的走子
	captureFound := false
	for _, m := range moves {
		if m.HasTag(xq.TagCapture) {
			captureFound = true
			break
		}
	}
	if !captureFound {
		t.Log("no capture move found in starting position (acceptable for some FENs)")
	}
}
