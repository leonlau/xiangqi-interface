package sdk

import (
	"fmt"
	"plugin"

	xq "github.com/leonlau/xiangqi-interface"
)

type SDK struct {
	engine xq.XiangqiEngine
}

func New(pluginPath string) (*SDK, error) {
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("open plugin %q: %w", pluginPath, err)
	}
	sym, err := p.Lookup("NewXiangqiEngine")
	if err != nil {
		return nil, fmt.Errorf("lookup NewXiangqiEngine: %w", err)
	}
	ctor, ok := sym.(func() xq.XiangqiEngine)
	if !ok {
		return nil, fmt.Errorf("symbol NewXiangqiEngine has type %T, want func() xiangqiinterface.XiangqiEngine", sym)
	}
	return &SDK{engine: ctor()}, nil
}

// === FEN ===

func (s *SDK) ValidateFEN(fen string) error            { return s.engine.ValidateFEN(fen) }
func (s *SDK) CanonicalFEN(fen string) (string, error) { return s.engine.CanonicalFEN(fen) }

// === Position queries ===

func (s *SDK) Turn(fen string) (xq.Color, error)                   { return s.engine.Turn(fen) }
func (s *SDK) InCheck(fen string, c xq.Color) (bool, error)        { return s.engine.InCheck(fen, c) }
func (s *SDK) Status(fen string) (xq.GameStatus, error)            { return s.engine.Status(fen) }
func (s *SDK) Board(fen string) (xq.Board, error)                  { return s.engine.Board(fen) }
func (s *SDK) Hash(fen string) (string, error)                     { return s.engine.Hash(fen) }
func (s *SDK) HalfMoveClock(fen string) (int, error)               { return s.engine.HalfMoveClock(fen) }
func (s *SDK) Rule60(fen string) (int, error)                      { return s.engine.Rule60(fen) }
func (s *SDK) Check10(fen string, c xq.Color) (int, error)         { return s.engine.Check10(fen, c) }
func (s *SDK) PieceMaterial(fen string) (xq.PieceMaterial, error)  { return s.engine.PieceMaterial(fen) }
func (s *SDK) BoardMap(fen string) (map[xq.Square]xq.Piece, error) { return s.engine.BoardMap(fen) }
func (s *SDK) BoardDraw(fen string) (string, error)                { return s.engine.BoardDraw(fen) }
func (s *SDK) BoardString(fen string) (string, error)              { return s.engine.BoardString(fen) }
func (s *SDK) BoardFlip(fen string, dir xq.FlipDirection) (xq.Board, error) {
	return s.engine.BoardFlip(fen, dir)
}
func (s *SDK) PositionString(fen string) (string, error) { return s.engine.PositionString(fen) }
func (s *SDK) PositionMarshalText(fen string) (string, error) {
	return s.engine.PositionMarshalText(fen)
}
func (s *SDK) PositionUnmarshalText(text string) (string, error) {
	return s.engine.PositionUnmarshalText(text)
}
func (s *SDK) PositionMarshalBinary(fen string) ([]byte, error) {
	return s.engine.PositionMarshalBinary(fen)
}
func (s *SDK) PositionUnmarshalBinary(data []byte) (string, error) {
	return s.engine.PositionUnmarshalBinary(data)
}

// === Move generation & state ===

func (s *SDK) ValidMoves(fen string) ([]xq.Move, error)        { return s.engine.ValidMoves(fen) }
func (s *SDK) ApplyMove(fen string, m xq.Move) (string, error) { return s.engine.ApplyMove(fen, m) }
func (s *SDK) ApplyMoves(fen string, moves []xq.Move) (string, error) {
	return s.engine.ApplyMoves(fen, moves)
}

// === Notation (multi-format) ===

func (s *SDK) DecodeMove(fen, notation string, n xq.Notation) (xq.Move, error) {
	return s.engine.DecodeMove(fen, notation, n)
}
func (s *SDK) EncodeMove(fen string, m xq.Move, n xq.Notation) (string, error) {
	return s.engine.EncodeMove(fen, m, n)
}

// === PGN ===

func (s *SDK) ParsePGN(pgn string) (xq.PGNGame, error)   { return s.engine.ParsePGN(pgn) }
func (s *SDK) EncodePGN(game xq.PGNGame) (string, error) { return s.engine.EncodePGN(game) }

// === Game (stateful) ===

func (s *SDK) NewGame(startFEN string, opts ...xq.GameOption) (xq.GameHandle, error) {
	return s.engine.NewGame(startFEN, opts...)
}
func (s *SDK) GameMove(h xq.GameHandle, m xq.Move) error           { return s.engine.GameMove(h, m) }
func (s *SDK) GameUndo(h xq.GameHandle) (xq.Move, error)           { return s.engine.GameUndo(h) }
func (s *SDK) GameUndoN(h xq.GameHandle, n int) ([]xq.Move, error) { return s.engine.GameUndoN(h, n) }
func (s *SDK) GameFEN(h xq.GameHandle) (string, error)             { return s.engine.GameFEN(h) }
func (s *SDK) GameMoves(h xq.GameHandle) ([]xq.Move, error)        { return s.engine.GameMoves(h) }
func (s *SDK) GamePositions(h xq.GameHandle) ([]string, error)     { return s.engine.GamePositions(h) }
func (s *SDK) GameComments(h xq.GameHandle) []string               { return s.engine.GameComments(h) }
func (s *SDK) GameStatus(h xq.GameHandle) (xq.GameStatus, error)   { return s.engine.GameStatus(h) }
func (s *SDK) GamePGN(h xq.GameHandle) (string, error)             { return s.engine.GamePGN(h) }
func (s *SDK) GameString(h xq.GameHandle) (string, error)          { return s.engine.GameString(h) }
func (s *SDK) GameMarshalText(h xq.GameHandle) (string, error)     { return s.engine.GameMarshalText(h) }
func (s *SDK) GameUnmarshalText(h xq.GameHandle, pgn string) error {
	return s.engine.GameUnmarshalText(h, pgn)
}
func (s *SDK) GameDraw(h xq.GameHandle, reason xq.DrawReason) error {
	return s.engine.GameDraw(h, reason)
}
func (s *SDK) GameResign(h xq.GameHandle, c xq.Color) error { return s.engine.GameResign(h, c) }
func (s *SDK) GameEligibleDraws(h xq.GameHandle) ([]xq.DrawReason, error) {
	return s.engine.GameEligibleDraws(h)
}
func (s *SDK) GameClone(h xq.GameHandle) (xq.GameHandle, error) { return s.engine.GameClone(h) }
func (s *SDK) GameSetNotation(h xq.GameHandle, n xq.Notation) error {
	return s.engine.GameSetNotation(h, n)
}
func (s *SDK) GameTagAdd(h xq.GameHandle, name, value string) (bool, error) {
	return s.engine.GameTagAdd(h, name, value)
}
func (s *SDK) GameTagGet(h xq.GameHandle, name string) (string, bool, error) {
	return s.engine.GameTagGet(h, name)
}
func (s *SDK) GameTagRemove(h xq.GameHandle, name string) (bool, error) {
	return s.engine.GameTagRemove(h, name)
}
func (s *SDK) GameJudge(h xq.GameHandle) (xq.Verdict, error) { return s.engine.GameJudge(h) }
func (s *SDK) GameMoveHistory(h xq.GameHandle) ([]xq.MoveHistory, error) {
	return s.engine.GameMoveHistory(h)
}
func (s *SDK) CloseGame(h xq.GameHandle) error { return s.engine.CloseGame(h) }

// === Opening book ===

func (s *SDK) OpeningFind(moves []xq.Move) (*xq.Opening, error) { return s.engine.OpeningFind(moves) }
func (s *SDK) OpeningFindExact(moves []xq.Move) (*xq.Opening, bool, error) {
	return s.engine.OpeningFindExact(moves)
}
func (s *SDK) OpeningPossible(moves []xq.Move) ([]xq.Opening, error) {
	return s.engine.OpeningPossible(moves)
}

// === UCI engine ===

func (s *SDK) NewEngine(path string) (xq.EngineHandle, error)        { return s.engine.NewEngine(path) }
func (s *SDK) EngineRun(h xq.EngineHandle, cmd xq.UCICmd) error      { return s.engine.EngineRun(h, cmd) }
func (s *SDK) EngineStop(h xq.EngineHandle) error                    { return s.engine.EngineStop(h) }
func (s *SDK) EngineClose(h xq.EngineHandle) error                   { return s.engine.EngineClose(h) }
func (s *SDK) EngineID(h xq.EngineHandle) (map[string]string, error) { return s.engine.EngineID(h) }
func (s *SDK) EngineOptions(h xq.EngineHandle) (map[string]xq.UCIOption, error) {
	return s.engine.EngineOptions(h)
}
func (s *SDK) EngineSearchResults(h xq.EngineHandle) (xq.UCISearchResults, error) {
	return s.engine.EngineSearchResults(h)
}
func (s *SDK) EngineKill(h xq.EngineHandle) error { return s.engine.EngineKill(h) }

// === ChessDB ===

func (s *SDK) NewChessDBClient(baseURL string) (xq.ChessDBClientHandle, error) {
	return s.engine.NewChessDBClient(baseURL)
}
func (s *SDK) CloseChessDBClient(h xq.ChessDBClientHandle) error {
	return s.engine.CloseChessDBClient(h)
}
func (s *SDK) ChessDBQuery(h xq.ChessDBClientHandle, board string, opts xq.ChessDBQueryOptions) ([]xq.ChessDBSuggestedMove, error) {
	return s.engine.ChessDBQuery(h, board, opts)
}
func (s *SDK) ChessDBQueryBest(h xq.ChessDBClientHandle, board string, opts xq.ChessDBQueryOptions) ([]xq.ChessDBSuggestedMove, error) {
	return s.engine.ChessDBQueryBest(h, board, opts)
}
func (s *SDK) ChessDBQueryPV(h xq.ChessDBClientHandle, board string, opts xq.ChessDBQueryOptions) (xq.ChessDBPV, error) {
	return s.engine.ChessDBQueryPV(h, board, opts)
}
func (s *SDK) ChessDBQueryScore(h xq.ChessDBClientHandle, board string, opts xq.ChessDBQueryOptions) (xq.ChessDBScore, error) {
	return s.engine.ChessDBQueryScore(h, board, opts)
}
func (s *SDK) ChessDBQuerySearch(h xq.ChessDBClientHandle, board string, opts xq.ChessDBQueryOptions) ([]xq.ChessDBSuggestedMove, error) {
	return s.engine.ChessDBQuerySearch(h, board, opts)
}
func (s *SDK) ChessDBQueryAll(h xq.ChessDBClientHandle, board string, opts xq.ChessDBQueryAllOptions) ([]xq.ChessDBMoveInfo, error) {
	return s.engine.ChessDBQueryAll(h, board, opts)
}
func (s *SDK) ChessDBQueryRule(h xq.ChessDBClientHandle, board string, movelist []string, opts xq.ChessDBQueryRuleOptions) ([]xq.ChessDBRuleResult, error) {
	return s.engine.ChessDBQueryRule(h, board, movelist, opts)
}
func (s *SDK) ChessDBQueue(h xq.ChessDBClientHandle, board string) error {
	return s.engine.ChessDBQueue(h, board)
}
func (s *SDK) ChessDBStore(h xq.ChessDBClientHandle, board, move string) error {
	return s.engine.ChessDBStore(h, board, move)
}

// === Zobrist ===

func (s *SDK) ZobristHash(fen string) (uint64, error) { return s.engine.ZobristHash(fen) }

// === Image ===

func (s *SDK) NewImageEncoder() (xq.ImageEncoderHandle, error) {
	return s.engine.NewImageEncoder()
}
func (s *SDK) CloseImageEncoder(h xq.ImageEncoderHandle) error {
	return s.engine.CloseImageEncoder(h)
}
func (s *SDK) ImageSVG(h xq.ImageEncoderHandle, fen string, opts ...xq.ImageOption) (string, error) {
	return s.engine.ImageSVG(h, fen, opts...)
}
