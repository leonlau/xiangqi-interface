# xiangqi-interface

Go Plugin 封装的象棋规则引擎 SDK。引入 `sdk/` 包并加载对应平台的 `.so`,即可使用走法生成、FEN、记谱、PGN、对局管理、开局库、子力评估等全部公开能力,无需了解插件加载细节,也不在 `go.mod` 中引入底层实现。

## 给合作方

### 1. 引入 SDK

```bash
go get github.com/leonlau/xiangqi-interface/sdk
go get github.com/leonlau/xiangqi-interface
```

### 2. 下载对应平台的 .so

到 [Releases](https://github.com/leonlau/xiangqi-interface/releases) 页面,按 `Go 版本 + 操作系统 + 架构` 选文件,例如 `libxiangqi-linux-amd64-go1.26.7.so`。**Go 版本必须严格对齐** —— Go plugin 要求主程序和 `.so` 用同一 Go 工具链构建。

### 3. 加载并使用

```go
package main

import (
    "fmt"
    "runtime"

    xqsdk "github.com/leonlau/xiangqi-interface/sdk"
    xq "github.com/leonlau/xiangqi-interface"
)

func main() {
    soName := fmt.Sprintf("./libxiangqi-%s-%s-go%s.so",
        runtime.GOOS, runtime.GOARCH, runtime.Version()[2:])
    sdk, err := xqsdk.New(soName)
    if err != nil { panic(err) }

    fen := "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"

    // FEN
    sdk.ValidateFEN(fen)
    canon, _ := sdk.CanonicalFEN(fen)

    // 局面查询
    turn, _ := sdk.Turn(fen)
    status, _ := sdk.Status(fen)
    inCheck, _ := sdk.InCheck(fen, xq.Red)
    board, _ := sdk.Board(fen)
    hash, _ := sdk.Hash(fen)
    hmc, _ := sdk.HalfMoveClock(fen)
    r60, _ := sdk.Rule60(fen)
    c10, _ := sdk.Check10(fen, xq.Red)
    mat, _ := sdk.PieceMaterial(fen)
    bmap, _ := sdk.BoardMap(fen)
    drawing, _ := sdk.BoardDraw(fen)

    // 走法
    moves, _ := sdk.ValidMoves(fen)
    nextFen, _ := sdk.ApplyMove(fen, moves[0])
    nextFens, _ := sdk.ApplyMoves(fen, moves[:3])

    // 多格式记谱
    m, _ := sdk.DecodeMove(fen, "炮二平五", xq.NotationChinese)
    uci, _ := sdk.EncodeMove(fen, m, xq.NotationUCI)
    cc, _ := sdk.EncodeMove(fen, m, xq.NotationChinese)
    cdb, _ := sdk.EncodeMove(fen, m, xq.NotationChessDB)
    fn, _ := sdk.EncodeMove(fen, m, xq.NotationFEN)
    if m.HasTag(xq.TagCapture) { /* 标记 */ }

    // PGN(带评注)
    pgn := `[Event "Test"]` + "\n\n" + `1. 炮二平五 { 好棋 } *`
    game, _ := sdk.ParsePGN(pgn)
    for i, m := range game.Moves {
        fmt.Println(m, game.Comments[i])
    }
    out, _ := sdk.EncodePGN(game)

    // 状态化对局
    h, _ := sdk.NewGame(fen,
        xq.UseRulesOpt{Rules: "asian"},
        xq.IgnoreRulesOpt{},
    )
    defer sdk.CloseGame(h)
    sdk.GameMove(h, m)
    undone, _ := sdk.GameUndo(h)
    multiUndone, _ := sdk.GameUndoN(h, 3)
    fmt.Println(sdk.GameFEN(h), sdk.GameStatus(h), sdk.GamePGN(h))
    fmt.Println(sdk.GameComments(h))
    sdk.GameDraw(h, xq.DrawBy50MoveRule)
    sdk.GameResign(h, xq.Red)

    // 开局库
    opening, _ := sdk.OpeningFind(moves)
    if opening != nil {
        fmt.Println(opening.Code, opening.Name)
    }

    // UCI 搜索引擎(需自带引擎二进制,如 fairy-stockfish)
    eng, _ := sdk.NewEngine("/usr/local/bin/fairy-stockfish")
    defer sdk.EngineClose(eng)
    sdk.EngineRun(eng, xq.UCICmdUCI{})           // 握手
    sdk.EngineRun(eng, xq.UCICmdIsReady{})
    sdk.EngineRun(eng, xq.UCICmdSetOption{Name: "Threads", Value: "4"})
    sdk.EngineRun(eng, xq.UCICmdPosition{StartFEN: fen})  // 起始局面
    sdk.EngineRun(eng, xq.UCICmdGo{MoveTime: 5 * time.Second})  // 搜 5 秒
    res, _ := sdk.EngineSearchResults(eng)
    fmt.Println("best move:", res.BestMove, "info:", res.Info)
}
```

## API 速查(62 个方法)

| 分类 | 方法 |
|---|---|
| **FEN** | `ValidateFEN` / `CanonicalFEN` |
| **Position queries** | `Turn` / `InCheck` / `Status` / `Board` / `Hash` / `HalfMoveClock` / `Rule60` / `Check10` / `PieceMaterial` / `BoardMap` / `BoardDraw` / `BoardString` / `BoardFlip` / `PositionString` / `PositionMarshalText` / `PositionUnmarshalText` / `PositionMarshalBinary` / `PositionUnmarshalBinary` |
| **Move** | `ValidMoves` / `ApplyMove` / `ApplyMoves` |
| **Notation** | `DecodeMove` / `EncodeMove` (4 种格式) |
| **PGN** | `ParsePGN` / `EncodePGN` |
| **Game(handle)** | `NewGame` / `GameMove` / `GameUndo` / `GameUndoN` / `GameFEN` / `GameMoves` / `GamePositions` / `GameComments` / `GameStatus` / `GamePGN` / `GameString` / `GameMarshalText` / `GameUnmarshalText` / `GameDraw` / `GameResign` / `GameEligibleDraws` / `GameClone` / `GameSetNotation` / `GameTagAdd` / `GameTagGet` / `GameTagRemove` / `GameJudge` / `GameMoveHistory` / `CloseGame` |
| **Opening** | `OpeningFind` / `OpeningFindExact` / `OpeningPossible` |
| **UCI engine** | `NewEngine` / `EngineRun` / `EngineStop` / `EngineClose` / `EngineID` / `EngineOptions` / `EngineSearchResults` |

> 注:`Position.UndoMove()` 不暴露 —— 它依赖 Position 内部 history,纯 FEN 输入下无效。用 `GameUndo` / `GameUndoN`(基于 handle)代替。

## 关键类型

```go
xq.Square        // File (a-i) + Rank (0-9)
xq.Piece         // Type (King/Advisor/Elephant/Horse/Rook/Cannon/Pawn) + Color
xq.Move          // S1 + S2 + Tags (Capture/Check 位掩码)
xq.Board         // [90]Piece  索引 = rank*9 + file
xq.GameStatus    // InProgress/Check/Checkmate/Stalemate/Draw
xq.Notation      // UCI/Chinese/ChessDB/FEN
xq.Opening       // Code (ECO) + Name + Variation
xq.GameHandle    // 不透明 int64 句柄
xq.GameOption    // UseRulesOpt / IgnoreRulesOpt / UseNotationOpt
xq.DrawReason    // Agreement/50Move/Repetition/Stalemate/InsufficientMaterial
xq.MoveTag       // TagCapture / TagCheck
xq.FlipDirection // FlipUpDown / FlipLeftRight
xq.Outcome       // Unknown / RedWon / BlackWon / Draw
xq.Method        // None/Checkmate/Resignation/DrawOffer/Stalemate/Repetition/...
xq.Verdict       // Outcome + Method + Reason(亚洲规则裁决)
xq.MoveHistory   // PreFEN + Move + PostFEN + Comments
xq.EngineHandle  // UCI 引擎进程句柄
xq.UCICmd        // sealed: UCICmdUCI/IsReady/SetOption/Position/Go/Stop/Quit/UCINewGame
xq.UCIInfo       // Depth/SelDepth/Time/Nodes/Score/PV/MultiPV
xq.UCISearchResults // BestMove + Ponder + Info
xq.UCIOption     // 引擎可配置选项
xq.UCIScore      // cp 或 mate 分数
```

## 限制

- 仅支持 Linux / macOS(Go plugin 限制)
- 合作方 `go.mod` 不会出现底层实现依赖,只见 SDK 包本身
- 跨进程 .so 必须与主程序 Go 版本一致

## 内部维护

`.github/workflows/build-plugin.yml` 跑了 4 runner × 2 Go 版本 = 8 个构建,产出 8 份 .so 并通过集成测试。tag `v*` 推送后自动发布为 GitHub Release。