package main

import (
	"os"
	"strconv"
	"sync"
	"testing"

	xq "github.com/leonlau/xiangqi-interface"
	xqsdk "github.com/leonlau/xiangqi-interface/sdk"
)

const pgnPath = "../data/xiangqi.pgn"

// undoBackTo 在 GameUndoN 测试里把局面回退到只保留这么多步的状态。
const undoBackTo = 10

var (
	pgnRawOnce sync.Once
	pgnRaw     string
	pgnRawErr  error
)

// pgnRawText 缓存 xiangqi.pgn 的全文。
func pgnRawText(t *testing.T) string {
	t.Helper()
	pgnRawOnce.Do(func() {
		data, err := os.ReadFile(pgnPath)
		if err != nil {
			pgnRawErr = err
			return
		}
		pgnRaw = string(data)
	})
	if pgnRawErr != nil {
		t.Fatalf("read %s: %v", pgnPath, pgnRawErr)
	}
	return pgnRaw
}

var (
	parsedOnce sync.Once
	parsedGame xq.PGNGame
	parsedErr  error
)

// parsedPGNGame 缓存 ParsePGN 的结果供多个测试复用。
func parsedPGNGame(t *testing.T) xq.PGNGame {
	t.Helper()
	parsedOnce.Do(func() {
		parsedGame, parsedErr = mustLoadSDKForData(t).ParsePGN(pgnRawText(t))
	})
	if parsedErr != nil {
		t.Fatalf("ParsePGN: %v", parsedErr)
	}
	return parsedGame
}

// === PGN 解析 / 编码 ===

func TestPGNData_ParsePGN(t *testing.T) {
	g := parsedPGNGame(t)

	if got := findTag(g.Tags, "Game"); got != "xiangqi" {
		t.Errorf("Game tag = %q, want xiangqi", got)
	}

	// xiangqi.pgn 一局 55 个半步(28 红 + 27 黑,最后黑未应子)
	if len(g.Moves) != 55 {
		t.Errorf("len(Moves) = %d, want 55", len(g.Moves))
	}

	// 中文 PGN 没有 Result tag,也不以 * / 1-0 等结尾,Result 是空字符串
	if g.Result != "" {
		t.Logf("Result = %q", g.Result)
	}
}

func TestPGNData_EncodePGN_RoundTrip(t *testing.T) {
	sdk := mustLoadSDKForData(t)
	g := parsedPGNGame(t)

	out, err := sdk.EncodePGN(g)
	if err != nil {
		t.Fatalf("EncodePGN: %v", err)
	}
	if out == "" {
		t.Fatal("EncodePGN 输出为空")
	}
	re, err := sdk.ParsePGN(out)
	if err != nil {
		t.Fatalf("ParsePGN(re-encoded): %v", err)
	}
	if got := findTag(re.Tags, "Game"); got != "xiangqi" {
		t.Errorf("re-encoded Game tag = %q, want xiangqi", got)
	}
	if len(re.Moves) != len(g.Moves) {
		t.Errorf("re-encoded len(Moves) = %d, want %d", len(re.Moves), len(g.Moves))
	}
}

// === Notation 走法 ===

func TestPGNData_EncodeMoveChinese(t *testing.T) {
	sdk := mustLoadSDKForData(t)
	g := parsedPGNGame(t)

	current := startFEN
	for i, m := range g.Moves {
		s, err := sdk.EncodeMove(current, m, xq.NotationChinese)
		if err != nil {
			t.Fatalf("EncodeMove[%d] Chinese: %v", i, err)
		}
		if s == "" {
			t.Errorf("EncodeMove[%d] Chinese 返回空字符串", i)
		}
		next, err := sdk.ApplyMove(current, m)
		if err != nil {
			t.Fatalf("ApplyMove[%d]=%+v: %v", i, m, err)
		}
		current = next
	}
}

func TestPGNData_NotationRoundTrip(t *testing.T) {
	sdk := mustLoadSDKForData(t)
	g := parsedPGNGame(t)

	current := startFEN
	for i, m := range g.Moves {
		// 走法 → 中文 → xq.Move → UCI 字符串 → xq.Move
		zh, err := sdk.EncodeMove(current, m, xq.NotationChinese)
		if err != nil {
			t.Fatalf("EncodeMove[%d] Chinese: %v", i, err)
		}
		decoded, err := sdk.DecodeMove(current, zh, xq.NotationChinese)
		if err != nil {
			t.Fatalf("DecodeMove[%d]=%q: %v", i, zh, err)
		}
		if decoded.S1 != m.S1 || decoded.S2 != m.S2 {
			t.Errorf("round-trip[%d]: %+v != %+v (Chinese=%q)", i, decoded, m, zh)
		}
		uci, err := sdk.EncodeMove(current, decoded, xq.NotationUCI)
		if err != nil {
			t.Fatalf("EncodeMove[%d] UCI: %v", i, err)
		}
		back, err := sdk.DecodeMove(current, uci, xq.NotationUCI)
		if err != nil {
			t.Fatalf("DecodeMove[%d]=%q UCI: %v", i, uci, err)
		}
		if back.S1 != m.S1 || back.S2 != m.S2 {
			t.Errorf("UCI round-trip[%d]: %+v != %+v", i, back, m)
		}
		current, _ = sdk.ApplyMove(current, m)
	}
}

// === Game handle 全生命周期 ===

func TestPGNData_GameNewAndApply(t *testing.T) {
	sdk := mustLoadSDKForData(t)
	g := parsedPGNGame(t)

	h, err := sdk.NewGame("")
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	defer sdk.CloseGame(h)

	if err := replayMoves(sdk, h, g.Moves); err != nil {
		t.Fatalf("replayMoves: %v", err)
	}
	gotMoves, err := sdk.GameMoves(h)
	if err != nil {
		t.Fatalf("GameMoves: %v", err)
	}
	if len(gotMoves) != len(g.Moves) {
		t.Errorf("GameMoves len = %d, want %d", len(gotMoves), len(g.Moves))
	}
	for i := range gotMoves {
		if gotMoves[i].S1 != g.Moves[i].S1 || gotMoves[i].S2 != g.Moves[i].S2 {
			t.Errorf("GameMoves[%d] = %+v, want %+v", i, gotMoves[i], g.Moves[i])
		}
	}
}

func TestPGNData_GameFEN(t *testing.T) {
	sdk := mustLoadSDKForData(t)
	g := parsedPGNGame(t)

	h, err := sdk.NewGame("")
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	defer sdk.CloseGame(h)

	current := startFEN
	for _, m := range g.Moves {
		current, _ = sdk.ApplyMove(current, m)
	}
	if err := replayMoves(sdk, h, g.Moves); err != nil {
		t.Fatalf("replayMoves: %v", err)
	}
	gotFEN, err := sdk.GameFEN(h)
	if err != nil {
		t.Fatalf("GameFEN: %v", err)
	}
	if gotFEN != current {
		t.Errorf("GameFEN 与 ApplyMove 链结果不一致")
	}
	t.Logf("最终 FEN: %s", gotFEN)
}

func TestPGNData_GamePGN(t *testing.T) {
	sdk := mustLoadSDKForData(t)
	g := parsedPGNGame(t)

	h, err := sdk.NewGame("")
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	defer sdk.CloseGame(h)

	if err := replayMoves(sdk, h, g.Moves); err != nil {
		t.Fatalf("replayMoves: %v", err)
	}

	gotPGN, err := sdk.GamePGN(h)
	if err != nil {
		t.Fatalf("GamePGN: %v", err)
	}
	if gotPGN == "" {
		t.Fatal("GamePGN 返回空")
	}
	gotStr, err := sdk.GameString(h)
	if err != nil {
		t.Fatalf("GameString: %v", err)
	}
	if gotPGN != gotStr {
		t.Errorf("GamePGN != GameString")
	}

	re, err := sdk.ParsePGN(gotPGN)
	if err != nil {
		t.Fatalf("ParsePGN(replayed): %v", err)
	}
	if len(re.Moves) != len(g.Moves) {
		t.Errorf("re-parsed len(Moves) = %d, want %d", len(re.Moves), len(g.Moves))
	}
}

func TestPGNData_GameMarshalText(t *testing.T) {
	sdk := mustLoadSDKForData(t)
	g := parsedPGNGame(t)

	h, err := sdk.NewGame("")
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	defer sdk.CloseGame(h)

	if err := replayMoves(sdk, h, g.Moves); err != nil {
		t.Fatalf("replayMoves: %v", err)
	}
	txt, err := sdk.GameMarshalText(h)
	if err != nil {
		t.Fatalf("GameMarshalText: %v", err)
	}
	if txt == "" {
		t.Fatal("GameMarshalText 返回空")
	}

	h2, err := sdk.NewGame("")
	if err != nil {
		t.Fatalf("NewGame(h2): %v", err)
	}
	defer sdk.CloseGame(h2)
	if err := sdk.GameUnmarshalText(h2, txt); err != nil {
		t.Fatalf("GameUnmarshalText: %v", err)
	}
	fen1, _ := sdk.GameFEN(h)
	fen2, _ := sdk.GameFEN(h2)
	if fen1 != fen2 {
		t.Errorf("Marshal/Unmarshal 文本后局面不一致:\n h1=%q\n h2=%q", fen1, fen2)
	}
}

func TestPGNData_GamePositions(t *testing.T) {
	sdk := mustLoadSDKForData(t)
	g := parsedPGNGame(t)

	h, err := sdk.NewGame("")
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	defer sdk.CloseGame(h)

	if err := replayMoves(sdk, h, g.Moves); err != nil {
		t.Fatalf("replayMoves: %v", err)
	}

	positions, err := sdk.GamePositions(h)
	if err != nil {
		t.Fatalf("GamePositions: %v", err)
	}
	if got, want := len(positions), len(g.Moves)+1; got != want {
		t.Errorf("GamePositions len = %d, want %d (起始 + 每步后)", got, want)
	}
	if positions[0] != startFEN {
		t.Errorf("positions[0] = %q, want startFEN", positions[0])
	}
	lastFEN, _ := sdk.GameFEN(h)
	if positions[len(positions)-1] != lastFEN {
		t.Error("positions 最后一项与 GameFEN 不一致")
	}
}

func TestPGNData_GameMoveHistory(t *testing.T) {
	sdk := mustLoadSDKForData(t)
	g := parsedPGNGame(t)

	h, err := sdk.NewGame("")
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	defer sdk.CloseGame(h)

	if err := replayMoves(sdk, h, g.Moves); err != nil {
		t.Fatalf("replayMoves: %v", err)
	}

	hist, err := sdk.GameMoveHistory(h)
	if err != nil {
		t.Fatalf("GameMoveHistory: %v", err)
	}
	if len(hist) != len(g.Moves) {
		t.Errorf("history len = %d, want %d", len(hist), len(g.Moves))
	}
	for i, mh := range hist {
		if mh.PreFEN == "" {
			t.Errorf("hist[%d].PreFEN 为空", i)
		}
		if mh.PostFEN == "" {
			t.Errorf("hist[%d].PostFEN 为空", i)
		}
		if mh.Move.S1 != g.Moves[i].S1 || mh.Move.S2 != g.Moves[i].S2 {
			t.Errorf("hist[%d].Move = %+v, want %+v", i, mh.Move, g.Moves[i])
		}
	}
}

func TestPGNData_GameEligibleDraws(t *testing.T) {
	sdk := mustLoadSDKForData(t)
	g := parsedPGNGame(t)

	h, err := sdk.NewGame("")
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	defer sdk.CloseGame(h)

	if err := replayMoves(sdk, h, g.Moves); err != nil {
		t.Fatalf("replayMoves: %v", err)
	}
	draws, err := sdk.GameEligibleDraws(h)
	if err != nil {
		t.Fatalf("GameEligibleDraws: %v", err)
	}
	t.Logf("可宣告和棋类型: %v", draws)
}

func TestPGNData_GameJudge(t *testing.T) {
	sdk := mustLoadSDKForData(t)
	g := parsedPGNGame(t)

	h, err := sdk.NewGame("")
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	defer sdk.CloseGame(h)

	if err := replayMoves(sdk, h, g.Moves); err != nil {
		t.Fatalf("replayMoves: %v", err)
	}
	v, err := sdk.GameJudge(h)
	if err != nil {
		t.Fatalf("GameJudge: %v", err)
	}
	t.Logf("裁判: outcome=%v method=%v reason=%q", v.Outcome, v.Method, v.Reason)
}

func TestPGNData_GameStatus(t *testing.T) {
	sdk := mustLoadSDKForData(t)
	g := parsedPGNGame(t)

	h, err := sdk.NewGame("")
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	defer sdk.CloseGame(h)

	if err := replayMoves(sdk, h, g.Moves); err != nil {
		t.Fatalf("replayMoves: %v", err)
	}
	st, err := sdk.GameStatus(h)
	if err != nil {
		t.Fatalf("GameStatus: %v", err)
	}
	t.Logf("局面状态: %v", st)
}

func TestPGNData_GameComments(t *testing.T) {
	sdk := mustLoadSDKForData(t)
	g := parsedPGNGame(t)

	h, err := sdk.NewGame("")
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	defer sdk.CloseGame(h)

	if err := replayMoves(sdk, h, g.Moves); err != nil {
		t.Fatalf("replayMoves: %v", err)
	}
	cs := sdk.GameComments(h)
	if len(cs) != len(g.Moves) {
		t.Errorf("GameComments len = %d, want %d (中文 PGN 无评注时全空字符串)", len(cs), len(g.Moves))
	}
}

// === Game 元信息 ===

func TestPGNData_GameClone_TagIsolation(t *testing.T) {
	sdk := mustLoadSDKForData(t)
	g := parsedPGNGame(t)

	h, err := sdk.NewGame("")
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	defer sdk.CloseGame(h)

	if err := replayMoves(sdk, h, g.Moves); err != nil {
		t.Fatalf("replayMoves: %v", err)
	}

	h2, err := sdk.GameClone(h)
	if err != nil {
		t.Fatalf("GameClone: %v", err)
	}
	defer sdk.CloseGame(h2)

	if h == h2 {
		t.Error("clone handle 与原 handle 相同")
	}
	fen1, _ := sdk.GameFEN(h)
	fen2, _ := sdk.GameFEN(h2)
	if fen1 != fen2 {
		t.Error("clone 后 FEN 与原局不一致")
	}
	if _, err := sdk.GameTagAdd(h2, "X-Test", "1"); err != nil {
		t.Fatalf("GameTagAdd on clone: %v", err)
	}
	if v, found, _ := sdk.GameTagGet(h, "X-Test"); found {
		t.Errorf("clone 上加的 tag 泄漏到原局: %q", v)
	}
}

func TestPGNData_GameTagCRUD(t *testing.T) {
	sdk := mustLoadSDKForData(t)

	h, err := sdk.NewGame("")
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	defer sdk.CloseGame(h)

	if _, err := sdk.GameTagAdd(h, "Site", "test"); err != nil {
		t.Fatalf("GameTagAdd: %v", err)
	}
	if v, found, err := sdk.GameTagGet(h, "Site"); err != nil || !found || v != "test" {
		t.Errorf("GameTagGet after Add: v=%q found=%v err=%v", v, found, err)
	}
	if _, err := sdk.GameTagAdd(h, "Site", "test2"); err != nil {
		t.Fatalf("GameTagAdd overwrite: %v", err)
	}
	if v, found, _ := sdk.GameTagGet(h, "Site"); !found || v != "test2" {
		t.Errorf("GameTagGet after overwrite: v=%q found=%v", v, found)
	}
	if _, err := sdk.GameTagRemove(h, "Site"); err != nil {
		t.Fatalf("GameTagRemove: %v", err)
	}
	if _, found, _ := sdk.GameTagGet(h, "Site"); found {
		t.Error("Site tag 移除后仍可读到")
	}
}

func TestPGNData_GameSetNotation(t *testing.T) {
	sdk := mustLoadSDKForData(t)
	g := parsedPGNGame(t)

	h, err := sdk.NewGame("")
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	defer sdk.CloseGame(h)

	if err := replayMoves(sdk, h, g.Moves); err != nil {
		t.Fatalf("replayMoves: %v", err)
	}

	for _, n := range []xq.Notation{xq.NotationUCI, xq.NotationChinese, xq.NotationChessDB, xq.NotationUCCI, xq.NotationFEN} {
		if err := sdk.GameSetNotation(h, n); err != nil {
			t.Errorf("GameSetNotation(%v): %v", n, err)
		}
	}
}

func TestPGNData_GameUndo(t *testing.T) {
	sdk := mustLoadSDKForData(t)
	g := parsedPGNGame(t)

	h, err := sdk.NewGame("")
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	defer sdk.CloseGame(h)

	if err := replayMoves(sdk, h, g.Moves); err != nil {
		t.Fatalf("replayMoves: %v", err)
	}

	last := g.Moves[len(g.Moves)-1]
	undone, err := sdk.GameUndo(h)
	if err != nil {
		t.Fatalf("GameUndo: %v", err)
	}
	if undone.S1 != last.S1 || undone.S2 != last.S2 {
		t.Errorf("GameUndo 返回 %+v, want %+v", undone, last)
	}
	left, _ := sdk.GameMoves(h)
	if len(left) != len(g.Moves)-1 {
		t.Errorf("undo 后 GameMoves len = %d, want %d", len(left), len(g.Moves)-1)
	}
}

func TestPGNData_GameUndoN(t *testing.T) {
	sdk := mustLoadSDKForData(t)
	g := parsedPGNGame(t)

	h, err := sdk.NewGame("")
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	defer sdk.CloseGame(h)

	if err := replayMoves(sdk, h, g.Moves); err != nil {
		t.Fatalf("replayMoves: %v", err)
	}

	n := len(g.Moves) - undoBackTo
	if n <= 0 {
		t.Skip("走法数太少")
	}
	undone, err := sdk.GameUndoN(h, n)
	if err != nil {
		t.Fatalf("GameUndoN: %v", err)
	}
	if len(undone) != n {
		t.Errorf("GameUndoN 返回 %d 步, want %d", len(undone), n)
	}
	left, _ := sdk.GameMoves(h)
	if len(left) != undoBackTo {
		t.Errorf("undo %d 步后剩余 = %d, want %d", n, len(left), undoBackTo)
	}
}

func TestPGNData_GameResignAndDraw(t *testing.T) {
	sdk := mustLoadSDKForData(t)
	g := parsedPGNGame(t)

	h, err := sdk.NewGame("")
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	defer sdk.CloseGame(h)

	if err := replayMoves(sdk, h, g.Moves); err != nil {
		t.Fatalf("replayMoves: %v", err)
	}

	if err := sdk.GameResign(h, xq.Black); err != nil {
		t.Fatalf("GameResign(Black): %v", err)
	}
	pgn, err := sdk.GamePGN(h)
	if err != nil {
		t.Fatalf("GamePGN after resign: %v", err)
	}
	if pgn == "" {
		t.Error("resign 后 GamePGN 为空")
	}

	// GameDraw(DrawByAgreement) 把和棋写进 Game;在另一盘重放同样走法验证可用
	h2, err := sdk.NewGame("")
	if err != nil {
		t.Fatalf("NewGame(h2): %v", err)
	}
	defer sdk.CloseGame(h2)
	if err := replayMoves(sdk, h2, g.Moves); err != nil {
		t.Fatalf("replayMoves(h2): %v", err)
	}
	if err := sdk.GameDraw(h2, xq.DrawByAgreement); err != nil {
		t.Errorf("GameDraw(DrawByAgreement): %v", err)
	}
}

// === 辅助 ===

func findTag(tags []xq.PGNTag, name string) string {
	for _, t := range tags {
		if t.Name == name {
			return t.Value
		}
	}
	return ""
}

func replayMoves(sdk *xqsdk.SDK, h xq.GameHandle, moves []xq.Move) error {
	for i, m := range moves {
		if err := sdk.GameMove(h, m); err != nil {
			return fmtErrAt(i, err)
		}
	}
	return nil
}

// fmtErrAt 给错误带上失败索引,避免调用方只知道"某一步错了"。
func fmtErrAt(i int, err error) error {
	if err == nil {
		return nil
	}
	return &moveErr{idx: i, err: err}
}

type moveErr struct {
	idx int
	err error
}

func (e *moveErr) Error() string {
	return "move " + strconv.Itoa(e.idx) + ": " + e.err.Error()
}
