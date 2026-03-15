package xiangqiinterface

// XiangqiEngine 是 plugin 必须实现的接口。SDK 通过 plugin.Lookup 拿到实现后,逐方法调用。
type XiangqiEngine interface {
	// === FEN ===
	ValidateFEN(fen string) error
	CanonicalFEN(fen string) (string, error)

	// === Position queries ===
	Turn(fen string) (Color, error)
	InCheck(fen string, c Color) (bool, error)
	Status(fen string) (GameStatus, error)
	Board(fen string) (Board, error)
	Hash(fen string) (string, error)
	HalfMoveClock(fen string) (int, error)
	Rule60(fen string) (int, error)
	Check10(fen string, c Color) (int, error)
	PieceMaterial(fen string) (PieceMaterial, error)
	BoardMap(fen string) (map[Square]Piece, error)
	BoardDraw(fen string) (string, error)
	BoardString(fen string) (string, error)
	BoardFlip(fen string, dir FlipDirection) (Board, error)
	PositionString(fen string) (string, error)
	PositionMarshalText(fen string) (string, error)
	PositionUnmarshalText(text string) (string, error)
	PositionMarshalBinary(fen string) ([]byte, error)
	PositionUnmarshalBinary(data []byte) (string, error)

	// === Move generation & state ===
	ValidMoves(fen string) ([]Move, error)
	ApplyMove(fen string, m Move) (string, error)
	ApplyMoves(fen string, moves []Move) (string, error)

	// === Notation (multi-format) ===
	DecodeMove(fen, notation string, n Notation) (Move, error)
	EncodeMove(fen string, m Move, n Notation) (string, error)

	// === PGN ===
	ParsePGN(pgn string) (PGNGame, error)
	EncodePGN(game PGNGame) (string, error)

	// === Game (stateful, via opaque handle) ===
	NewGame(startFEN string, opts ...GameOption) (GameHandle, error)
	GameMove(h GameHandle, m Move) error
	GameUndo(h GameHandle) (Move, error)
	GameUndoN(h GameHandle, n int) ([]Move, error)
	GameFEN(h GameHandle) (string, error)
	GameMoves(h GameHandle) ([]Move, error)
	GamePositions(h GameHandle) ([]string, error)
	GameComments(h GameHandle) []string
	GameStatus(h GameHandle) (GameStatus, error)
	GamePGN(h GameHandle) (string, error)
	GameString(h GameHandle) (string, error) // 同 GamePGN
	GameMarshalText(h GameHandle) (string, error)
	GameUnmarshalText(h GameHandle, pgn string) error
	GameDraw(h GameHandle, reason DrawReason) error
	GameResign(h GameHandle, c Color) error
	GameEligibleDraws(h GameHandle) ([]DrawReason, error)
	GameClone(h GameHandle) (GameHandle, error)
	GameSetNotation(h GameHandle, n Notation) error
	GameTagAdd(h GameHandle, name, value string) (bool, error)
	GameTagGet(h GameHandle, name string) (string, bool, error)
	GameTagRemove(h GameHandle, name string) (bool, error)
	GameJudge(h GameHandle) (Verdict, error)
	GameMoveHistory(h GameHandle) ([]MoveHistory, error)
	CloseGame(h GameHandle) error

	// === Opening book ===
	OpeningFind(moves []Move) (*Opening, error)
	OpeningFindExact(moves []Move) (*Opening, bool, error)
	OpeningPossible(moves []Move) ([]Opening, error)

	// === UCI engine ===
	NewEngine(path string) (EngineHandle, error)
	EngineRun(h EngineHandle, cmd UCICmd) error
	EngineStop(h EngineHandle) error
	EngineClose(h EngineHandle) error
	EngineID(h EngineHandle) (map[string]string, error)
	EngineOptions(h EngineHandle) (map[string]UCIOption, error)
	EngineSearchResults(h EngineHandle) (UCISearchResults, error)

	// === UCI 引擎 ===

	// EngineKill 不发送 CmdQuit,直接终止引擎进程,适合 graceful shutdown 失败的场景。
	EngineKill(h EngineHandle) error

	// === ChessDB 云库 ===

	// NewChessDBClient baseURL 为空使用 chessdb.cn 默认端点。
	NewChessDBClient(baseURL string) (ChessDBClientHandle, error)
	CloseChessDBClient(h ChessDBClientHandle) error

	ChessDBQuery(h ChessDBClientHandle, board string, opts ChessDBQueryOptions) ([]ChessDBSuggestedMove, error)
	ChessDBQueryBest(h ChessDBClientHandle, board string, opts ChessDBQueryOptions) ([]ChessDBSuggestedMove, error)
	ChessDBQueryPV(h ChessDBClientHandle, board string, opts ChessDBQueryOptions) (ChessDBPV, error)
	ChessDBQueryScore(h ChessDBClientHandle, board string, opts ChessDBQueryOptions) (ChessDBScore, error)
	ChessDBQuerySearch(h ChessDBClientHandle, board string, opts ChessDBQueryOptions) ([]ChessDBSuggestedMove, error)
	ChessDBQueryAll(h ChessDBClientHandle, board string, opts ChessDBQueryAllOptions) ([]ChessDBMoveInfo, error)
	ChessDBQueryRule(h ChessDBClientHandle, board string, movelist []string, opts ChessDBQueryRuleOptions) ([]ChessDBRuleResult, error)
	ChessDBQueue(h ChessDBClientHandle, board string) error
	ChessDBStore(h ChessDBClientHandle, board, move string) error

	// === Zobrist ===

	// ZobristHash 计算 Fairy-Stockfish 兼容的 64 位 Zobrist 键。
	ZobristHash(fen string) (uint64, error)

	// === Image ===

	NewImageEncoder() (ImageEncoderHandle, error)
	CloseImageEncoder(h ImageEncoderHandle) error
	ImageSVG(h ImageEncoderHandle, fen string, opts ...ImageOption) (string, error)
}
