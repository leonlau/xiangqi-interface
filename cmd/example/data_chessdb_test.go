package main

import (
	"bufio"
	"os"
	"strings"
	"sync"
	"testing"

	xq "github.com/leonlau/xiangqi-interface"
	xqsdk "github.com/leonlau/xiangqi-interface/sdk"
)

// board 一边 9 格;索引写法 r*9+f。
const boardSide = 9

// fullPieceCount 起始局面红 + 黑合计 32 子。
const fullPieceCount = 32

// chessdbMoveLen ChessDB 走法串固定 4 字符(源/目,源/目各占 2)。
const chessdbMoveLen = 4

// openingPrefixLen 开局识别测试用的走法前缀长度(短到能命中 ECO 常见布局)。
const openingPrefixLen = 3

func idx(r, f int) int { return r*boardSide + f }

// chessdbEntry 是 chessdb.txt 单行的解析结果。
type chessdbEntry struct {
	startFEN string   // 补齐 move counters 后可用 ValidateFEN 等方法
	moves    []string // ChessDB 走法串,如 "h0i2"
}

var (
	chessdbOnce     sync.Once
	chessdbEntries  []chessdbEntry
	chessdbEntriesE error
)

func loadChessDBData(t *testing.T) []chessdbEntry {
	t.Helper()
	chessdbOnce.Do(func() {
		f, err := os.Open("../data/chessdb.txt")
		if err != nil {
			chessdbEntriesE = err
			return
		}
		defer f.Close()

		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.Fields(line)
			if len(parts) < 4 || parts[1] != "w" || parts[2] != "moves" {
				chessdbEntriesE = errParseFormat(line)
				return
			}
			chessdbEntries = append(chessdbEntries, chessdbEntry{
				startFEN: parts[0] + " " + parts[1] + " - - 0 1",
				moves:    append([]string(nil), parts[3:]...),
			})
		}
		if err := sc.Err(); err != nil {
			chessdbEntriesE = err
		}
	})
	if chessdbEntriesE != nil {
		t.Fatalf("load chessdb.txt: %v", chessdbEntriesE)
	}
	return chessdbEntries
}

// errParseFormat 避免在 Do 内直接构造 fmt.Errorf 带来的 import 噪音。
func errParseFormat(line string) error {
	return &parseError{line: line}
}

type parseError struct{ line string }

func (e *parseError) Error() string { return "chessdb.txt 行格式不符合预期: " + e.line }

func firstEntry(t *testing.T, entries []chessdbEntry) chessdbEntry {
	t.Helper()
	if len(entries) == 0 {
		t.Fatal("chessdb.txt 没有 entry")
	}
	return entries[0]
}

var (
	sdkOnce sync.Once
	sdk     *xqsdk.SDK
	sdkErr  error
)

// mustLoadSDKForData 加载本地构建的 .so 并缓存;无 .so 时跳过整个测试。
func mustLoadSDKForData(t *testing.T) *xqsdk.SDK {
	t.Helper()
	sdkOnce.Do(func() {
		path := os.Getenv("ENGINE_PLUGIN")
		if path == "" {
			path = "../dist/libengine.so"
		}
		sdk, sdkErr = xqsdk.New(path)
	})
	if sdkErr != nil {
		t.Skipf("plugin 未构建 (设置 ENGINE_PLUGIN 后重试): %v", sdkErr)
	}
	return sdk
}

// === 解析 ===

func TestChessDBData_ParseFile(t *testing.T) {
	entries := loadChessDBData(t)
	if len(entries) == 0 {
		t.Fatal("chessdb.txt 解析得到 0 条 entry")
	}
	e := firstEntry(t, entries)
	wantFEN := "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"
	if e.startFEN != wantFEN {
		t.Errorf("startFEN = %q, want %q", e.startFEN, wantFEN)
	}
	if len(e.moves) == 0 {
		t.Error("moves 为空")
	}
	for i, m := range e.moves {
		if len(m) != chessdbMoveLen {
			t.Errorf("moves[%d] = %q 长度不是 %d (ChessDB 走法形如 h0i2)", i, m, chessdbMoveLen)
		}
	}
}

// === 起始 FEN 上的 Position / Move 系列 ===

func TestChessDBData_StartFEN_Queries(t *testing.T) {
	sdk := mustLoadSDKForData(t)
	e := firstEntry(t, loadChessDBData(t))

	if err := sdk.ValidateFEN(e.startFEN); err != nil {
		t.Fatalf("ValidateFEN: %v", err)
	}
	canon, err := sdk.CanonicalFEN(e.startFEN)
	if err != nil {
		t.Fatalf("CanonicalFEN: %v", err)
	}
	if canon != e.startFEN {
		t.Errorf("CanonicalFEN = %q, want %q", canon, e.startFEN)
	}

	turn, err := sdk.Turn(e.startFEN)
	if err != nil {
		t.Fatalf("Turn: %v", err)
	}
	if turn != xq.Red {
		t.Errorf("Turn = %v, want Red", turn)
	}

	hash1, err := sdk.Hash(e.startFEN)
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	if len(hash1) != 32 {
		t.Errorf("Hash length = %d, want 32", len(hash1))
	}
	hash2, err := sdk.Hash(e.startFEN)
	if err != nil {
		t.Fatalf("Hash(2): %v", err)
	}
	if hash1 != hash2 {
		t.Error("Hash 不稳定")
	}

	board, err := sdk.Board(e.startFEN)
	if err != nil {
		t.Fatalf("Board: %v", err)
	}
	occupied := 0
	for _, p := range board {
		if p.Type != xq.NoPieceType {
			occupied++
		}
	}
	if occupied != fullPieceCount {
		t.Errorf("起始局面 occupied = %d, want %d", occupied, fullPieceCount)
	}

	bm, err := sdk.BoardMap(e.startFEN)
	if err != nil {
		t.Fatalf("BoardMap: %v", err)
	}
	if len(bm) != fullPieceCount {
		t.Errorf("BoardMap len = %d, want %d", len(bm), fullPieceCount)
	}

	if _, err := sdk.BoardDraw(e.startFEN); err != nil {
		t.Errorf("BoardDraw: %v", err)
	}
	if _, err := sdk.BoardString(e.startFEN); err != nil {
		t.Errorf("BoardString: %v", err)
	}
	if _, err := sdk.PieceMaterial(e.startFEN); err != nil {
		t.Errorf("PieceMaterial: %v", err)
	}
	if n, err := sdk.HalfMoveClock(e.startFEN); err != nil || n != 0 {
		t.Errorf("HalfMoveClock = %d, err = %v, want 0", n, err)
	}
	if n, err := sdk.Rule60(e.startFEN); err != nil || n != 0 {
		t.Errorf("Rule60 = %d, err = %v, want 0", n, err)
	}

	updown, err := sdk.BoardFlip(e.startFEN, xq.FlipUpDown)
	if err != nil {
		t.Fatalf("BoardFlip UpDown: %v", err)
	}
	// 红方底线 rank 0 上移到 rank 9
	if p := updown[idx(9, 0)]; p.Type != xq.Rook || p.Color != xq.Red {
		t.Errorf("BoardFlip UpDown 后 board[9][0] = %v, want Red Rook", p)
	}
	leftright, err := sdk.BoardFlip(e.startFEN, xq.FlipLeftRight)
	if err != nil {
		t.Fatalf("BoardFlip LeftRight: %v", err)
	}
	// 红方 a 列(Rook) -> i 列
	if p := leftright[idx(0, 8)]; p.Type != xq.Rook || p.Color != xq.Red {
		t.Errorf("BoardFlip LeftRight 后 board[0][8] = %v, want Red Rook", p)
	}
}

// === ChessDB 走法解析与往返 ===

func TestChessDBData_DecodeChessDBMoves(t *testing.T) {
	sdk := mustLoadSDKForData(t)
	e := firstEntry(t, loadChessDBData(t))

	current := e.startFEN
	for i, m := range e.moves {
		decoded, err := sdk.DecodeMove(current, m, xq.NotationChessDB)
		if err != nil {
			t.Fatalf("DecodeMove[%d]=%q: %v", i, m, err)
		}
		next, err := sdk.ApplyMove(current, decoded)
		if err != nil {
			t.Fatalf("ApplyMove[%d]=%+v: %v", i, decoded, err)
		}
		if next == current {
			t.Errorf("第 %d 步 (%q) 走子后 FEN 未变化", i, m)
		}
		current = next
	}
	// 起始局面的合法性查询
	if _, err := sdk.ValidMoves(e.startFEN); err != nil {
		t.Errorf("ValidMoves: %v", err)
	}
}

func TestChessDBData_EncodeMoveChessDB(t *testing.T) {
	sdk := mustLoadSDKForData(t)
	e := firstEntry(t, loadChessDBData(t))

	current := e.startFEN
	for i, m := range e.moves {
		decoded, err := sdk.DecodeMove(current, m, xq.NotationChessDB)
		if err != nil {
			t.Fatalf("DecodeMove[%d]=%q: %v", i, m, err)
		}
		got, err := sdk.EncodeMove(current, decoded, xq.NotationChessDB)
		if err != nil {
			t.Fatalf("EncodeMove[%d]: %v", i, err)
		}
		if got != m {
			t.Errorf("EncodeMove[%d] = %q, want %q", i, got, m)
		}
		next, err := sdk.ApplyMove(current, decoded)
		if err != nil {
			t.Fatalf("ApplyMove[%d]: %v", i, err)
		}
		current = next
	}
}

// === 一次性走完所有走法 ===

func TestChessDBData_ApplyMoves(t *testing.T) {
	sdk := mustLoadSDKForData(t)
	e := firstEntry(t, loadChessDBData(t))

	decoded := make([]xq.Move, 0, len(e.moves))
	current := e.startFEN
	for i, m := range e.moves {
		d, err := sdk.DecodeMove(current, m, xq.NotationChessDB)
		if err != nil {
			t.Fatalf("DecodeMove[%d]=%q: %v", i, m, err)
		}
		decoded = append(decoded, d)
		current, _ = sdk.ApplyMove(current, d)
	}

	batch, err := sdk.ApplyMoves(e.startFEN, decoded)
	if err != nil {
		t.Fatalf("ApplyMoves: %v", err)
	}
	if batch != current {
		t.Errorf("ApplyMoves 与逐步 ApplyMove 结果不一致:\n batch=%q\n step =%q", batch, current)
	}

	if _, err := sdk.Status(current); err != nil {
		t.Errorf("Status: %v", err)
	}

	for _, c := range []xq.Color{xq.Red, xq.Black} {
		if _, err := sdk.InCheck(current, c); err != nil {
			t.Errorf("InCheck(%v): %v", c, err)
		}
		if _, err := sdk.Check10(current, c); err != nil {
			t.Errorf("Check10(%v): %v", c, err)
		}
	}
	for name, fn := range map[string]func() (any, error){
		"PositionString": func() (any, error) { _, e := sdk.PositionString(current); return nil, e },
		"BoardString":    func() (any, error) { _, e := sdk.BoardString(current); return nil, e },
		"BoardDraw":      func() (any, error) { _, e := sdk.BoardDraw(current); return nil, e },
		"BoardMap":       func() (any, error) { _, e := sdk.BoardMap(current); return nil, e },
		"PieceMaterial":  func() (any, error) { _, e := sdk.PieceMaterial(current); return nil, e },
	} {
		if _, err := fn(); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
	if n, err := sdk.HalfMoveClock(current); err != nil || n < 0 {
		t.Errorf("HalfMoveClock = %d, err = %v", n, err)
	}
	if n, err := sdk.Rule60(current); err != nil || n < 0 {
		t.Errorf("Rule60 = %d, err = %v", n, err)
	}
}

// === Position 序列化 round-trip ===

func TestChessDBData_PositionMarshal(t *testing.T) {
	sdk := mustLoadSDKForData(t)
	e := firstEntry(t, loadChessDBData(t))

	// 走 3 步给非平凡的局面
	current := e.startFEN
	for i := 0; i < openingPrefixLen && i < len(e.moves); i++ {
		d, err := sdk.DecodeMove(current, e.moves[i], xq.NotationChessDB)
		if err != nil {
			t.Fatalf("DecodeMove[%d]: %v", i, err)
		}
		next, err := sdk.ApplyMove(current, d)
		if err != nil {
			t.Fatalf("ApplyMove[%d]: %v", i, err)
		}
		current = next
	}

	text, err := sdk.PositionMarshalText(current)
	if err != nil {
		t.Fatalf("PositionMarshalText: %v", err)
	}
	roundText, err := sdk.PositionUnmarshalText(text)
	if err != nil {
		t.Fatalf("PositionUnmarshalText: %v", err)
	}
	if roundText != current {
		t.Errorf("Text round-trip 不一致:\n got=%q\n want=%q", roundText, current)
	}

	bin, err := sdk.PositionMarshalBinary(current)
	if err != nil {
		t.Fatalf("PositionMarshalBinary: %v", err)
	}
	if len(bin) == 0 {
		t.Error("MarshalBinary 返回空字节切片")
	}
	roundBin, err := sdk.PositionUnmarshalBinary(bin)
	if err != nil {
		t.Fatalf("PositionUnmarshalBinary: %v", err)
	}
	if roundBin != current {
		t.Errorf("Binary round-trip 不一致:\n got=%q\n want=%q", roundBin, current)
	}

	if _, err := sdk.ValidMoves(current); err != nil {
		t.Errorf("ValidMoves on moved position: %v", err)
	}
}

// === 开局识别 ===

// entryXqMoves 缓存 chessdb.txt 第一条 entry 走法对应的 xq.Move 列表。
var (
	entryXqOnce sync.Once
	entryXqMoves []xq.Move
)

func chessdbEntryToXqMoves(t *testing.T, sdk *xqsdk.SDK, e chessdbEntry) []xq.Move {
	t.Helper()
	entryXqOnce.Do(func() {
		var out []xq.Move
		current := e.startFEN
		for i, m := range e.moves {
			d, err := sdk.DecodeMove(current, m, xq.NotationChessDB)
			if err != nil {
				t.Fatalf("DecodeMove[%d]=%q: %v", i, m, err)
			}
			out = append(out, d)
			current, _ = sdk.ApplyMove(current, d)
		}
		entryXqMoves = out
	})
	return entryXqMoves
}

func TestChessDBData_OpeningFind(t *testing.T) {
	sdk := mustLoadSDKForData(t)
	e := firstEntry(t, loadChessDBData(t))
	moves := chessdbEntryToXqMoves(t, sdk, e)
	if len(moves) > openingPrefixLen {
		moves = moves[:openingPrefixLen]
	}
	o, err := sdk.OpeningFind(moves)
	if err != nil {
		t.Fatalf("OpeningFind: %v", err)
	}
	if o != nil {
		t.Logf("OpeningFind 命中: code=%s name=%s variation=%q", o.Code, o.Name, o.Variation)
	}
}

func TestChessDBData_OpeningFindExact(t *testing.T) {
	sdk := mustLoadSDKForData(t)
	e := firstEntry(t, loadChessDBData(t))
	moves := chessdbEntryToXqMoves(t, sdk, e)
	o, exact, err := sdk.OpeningFindExact(moves)
	if err != nil {
		t.Fatalf("OpeningFindExact: %v", err)
	}
	if o != nil {
		t.Logf("OpeningFindExact 命中: code=%s name=%s exact=%v", o.Code, o.Name, exact)
	}
}

func TestChessDBData_OpeningPossible(t *testing.T) {
	sdk := mustLoadSDKForData(t)
	e := firstEntry(t, loadChessDBData(t))
	moves := chessdbEntryToXqMoves(t, sdk, e)
	if len(moves) > openingPrefixLen {
		moves = moves[:openingPrefixLen]
	}
	os, err := sdk.OpeningPossible(moves)
	if err != nil {
		t.Fatalf("OpeningPossible: %v", err)
	}
	t.Logf("OpeningPossible 在 %d 步前缀后命中 %d 条候选", len(moves), len(os))
	for _, o := range os {
		t.Logf("  - %s %s", o.Code, o.Name)
	}
}
