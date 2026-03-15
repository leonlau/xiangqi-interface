// Command example 加载本地构建的 .so 并跑一套 SDK smoke test，
// 用于在 CI 编译 .so 完成后做端到端验证。
//
// 用法:
//   ENGINE_PLUGIN=../dist/libengine.so CGO_ENABLED=1 go run .
// 或在 CI:
//   cd cmd/example && ENGINE_PLUGIN=../dist/libengine-...so CGO_ENABLED=1 go run .
package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	xq "github.com/leonlau/xiangqi-interface"
	xqsdk "github.com/leonlau/xiangqi-interface/sdk"
)

const startFEN = "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "FAIL:", err)
		os.Exit(1)
	}
}

func run() error {
	path := os.Getenv("ENGINE_PLUGIN")
	if path == "" {
		path = "../dist/libengine.so"
	}
	sdk, err := xqsdk.New(path)
	if err != nil {
		return fmt.Errorf("load %s: %w", path, err)
	}

	fmt.Println()
	fmt.Println("==== xiangqi-interface SDK smoke test ====")
	fmt.Printf("Go:   %s\n", runtime.Version()[2:])
	fmt.Printf("OS:   %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf(".so:  %s\n", path)
	fmt.Println()

	fmt.Println("[1/6] FEN")
	if err := sdk.ValidateFEN(startFEN); err != nil {
		return fmt.Errorf("ValidateFEN: %w", err)
	}
	canon, _ := sdk.CanonicalFEN(startFEN)
	fmt.Printf("      CanonicalFEN: %s\n", canon)

	fmt.Println("[2/6] 走法生成")
	moves, err := sdk.ValidMoves(startFEN)
	if err != nil {
		return fmt.Errorf("ValidMoves: %w", err)
	}
	if len(moves) != 44 {
		return fmt.Errorf("起始合法走法数: got %d want 44", len(moves))
	}
	fmt.Printf("      起始合法走法: %d 手\n", len(moves))

	fmt.Println("[3/6] 记谱")
	m, err := sdk.DecodeMove(startFEN, "炮二平五", xq.NotationChinese)
	if err != nil {
		return fmt.Errorf("DecodeMove: %w", err)
	}
	for _, n := range []xq.Notation{xq.NotationUCI, xq.NotationChinese, xq.NotationChessDB, xq.NotationFEN} {
		s, err := sdk.EncodeMove(startFEN, m, n)
		if err != nil {
			return fmt.Errorf("EncodeMove %v: %w", n, err)
		}
		fmt.Printf("      %-9s %s\n", n, s)
	}

	fmt.Println("[4/6] PGN")
	pgn, err := sdk.EncodePGN(xq.PGNGame{Moves: []xq.Move{m}})
	if err != nil {
		return fmt.Errorf("EncodePGN: %w", err)
	}
	fmt.Printf("      %s\n", strings.TrimSpace(pgn))

	fmt.Println("[5/6] Game handle")
	h, err := sdk.NewGame(startFEN, xq.UseRulesOpt{Rules: "asian"}, xq.IgnoreRulesOpt{})
	if err != nil {
		return fmt.Errorf("NewGame: %w", err)
	}
	defer sdk.CloseGame(h)
	if err := sdk.GameMove(h, m); err != nil {
		return fmt.Errorf("GameMove: %w", err)
	}
	fen, _ := sdk.GameFEN(h)
	st, _ := sdk.GameStatus(h)
	fmt.Printf("      走完「炮二平五」: %s\n", fen)
	fmt.Printf("      状态: %v\n", st)
	if op, err := sdk.OpeningFind([]xq.Move{m}); err == nil && op != nil {
		fmt.Printf("      起始第一步命中: %s — %s\n", op.Code, op.Name)
	}

	fmt.Println("[6/6] UCI 引擎")
	bin := os.Getenv("UCI_ENGINE")
	if bin == "" {
		bin = "/usr/local/bin/fairy-stockfish"
	}
	eng, err := sdk.NewEngine(bin)
	if err != nil {
		fmt.Printf("      跳过（%s 不可用: %v）\n", bin, err)
		fmt.Println()
		fmt.Println("==== 全部通过 ====")
		fmt.Println()
		return nil
	}
	defer sdk.EngineClose(eng)
	for _, cmd := range []xq.UCICmd{xq.UCICmdUCI{}, xq.UCICmdIsReady{}, xq.UCICmdPosition{StartFEN: startFEN}, xq.UCICmdGo{MoveTime: 300 * time.Millisecond}} {
		if err := sdk.EngineRun(eng, cmd); err != nil {
			return fmt.Errorf("EngineRun %T: %w", cmd, err)
		}
	}
	res, err := sdk.EngineSearchResults(eng)
	if err != nil {
		return fmt.Errorf("EngineSearchResults: %w", err)
	}
	fmt.Printf("      bestmove: %v\n", res.BestMove)

	fmt.Println()
	fmt.Println("==== 全部通过 ====")
	fmt.Println()
	return nil
}