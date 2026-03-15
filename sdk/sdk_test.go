package sdk

import (
	"errors"
	"testing"

	xq "github.com/leonlau/xiangqi-interface"
)

type fakeEngine struct {
	validateFENErr  error
	canonicalFENOut string
	canonicalFENErr error

	turnOut                    xq.Color
	turnErr                    error
	inCheck                    bool
	inCheckErr                 error
	statusOut                  xq.GameStatus
	statusErr                  error
	boardOut                   xq.Board
	boardErr                   error
	hashOut                    string
	hashErr                    error
	halfMoveOut                int
	halfMoveErr                error
	rule60Out                  int
	rule60Err                  error
	check10Out                 int
	check10Err                 error
	pieceMaterialOut           xq.PieceMaterial
	pieceMaterialErr           error
	boardMapOut                map[xq.Square]xq.Piece
	boardMapErr                error
	boardDrawOut               string
	boardDrawErr               error
	boardStringOut             string
	boardStringErr             error
	boardFlipOut               xq.Board
	boardFlipErr               error
	positionStringOut          string
	positionStringErr          error
	positionMarshalTextOut     string
	positionMarshalTextErr     error
	positionUnmarshalTextOut   string
	positionUnmarshalTextErr   error
	positionMarshalBinaryOut   []byte
	positionMarshalBinaryErr   error
	positionUnmarshalBinaryOut string
	positionUnmarshalBinaryErr error

	validMovesOut []xq.Move
	validMovesErr error

	applyMoveOut  string
	applyMoveErr  error
	applyMovesOut string
	applyMovesErr error

	decodeMoveOut xq.Move
	decodeMoveErr error
	encodeMoveOut string
	encodeMoveErr error

	parsePGNOut  xq.PGNGame
	parsePGNErr  error
	encodePGNOut string
	encodePGNErr error

	newGameOut           xq.GameHandle
	newGameErr           error
	gameMoveErr          error
	gameUndoOut          xq.Move
	gameUndoErr          error
	gameUndoNOut         []xq.Move
	gameUndoNErr         error
	gameFENOut           string
	gameFENErr           error
	gameMovesOut         []xq.Move
	gameMovesErr         error
	gamePositionsOut     []string
	gamePositionsErr     error
	gameCommentsOut      []string
	gameStatusOut        xq.GameStatus
	gameStatusErr        error
	gamePGNOut           string
	gamePGNErr           error
	gameStringOut        string
	gameStringErr        error
	gameMarshalTextOut   string
	gameMarshalTextErr   error
	gameUnmarshalTextErr error
	gameDrawErr          error
	gameResignErr        error
	gameEligibleDrawsOut []xq.DrawReason
	gameEligibleDrawsErr error
	gameCloneOut         xq.GameHandle
	gameCloneErr         error
	gameSetNotationErr   error
	gameTagAddOut        bool
	gameTagAddErr        error
	gameTagGetOut        string
	gameTagGetFound      bool
	gameTagGetErr        error
	gameTagRemoveOut     bool
	gameTagRemoveErr     error
	gameJudgeOut         xq.Verdict
	gameJudgeErr         error
	gameMoveHistoryOut   []xq.MoveHistory
	gameMoveHistoryErr   error
	closeGameErr         error

	openingFindOut        *xq.Opening
	openingFindErr        error
	openingFindExactOut   *xq.Opening
	openingFindExactFound bool
	openingFindExactErr   error
	openingPossibleOut    []xq.Opening
	openingPossibleErr    error

	newEngineOut           xq.EngineHandle
	newEngineErr           error
	engineRunErr           error
	engineStopErr          error
	engineCloseErr         error
	engineIDOut            map[string]string
	engineIDErr            error
	engineOptionsOut       map[string]xq.UCIOption
	engineOptionsErr       error
	engineSearchResultsOut xq.UCISearchResults
	engineSearchResultsErr error

	// 新增 16 个方法对应的字段
	engineKillErr         error
	newChessDBClientErr   error
	closeChessDBClientErr error
	chessDBQueryOut       []xq.ChessDBSuggestedMove
	chessDBQueryErr       error
	chessDBQueryBestOut   []xq.ChessDBSuggestedMove
	chessDBQueryBestErr   error
	chessDBQueryPVOut     xq.ChessDBPV
	chessDBQueryPVErr     error
	chessDBQueryScoreOut  xq.ChessDBScore
	chessDBQueryScoreErr  error
	chessDBQuerySearchOut []xq.ChessDBSuggestedMove
	chessDBQuerySearchErr error
	chessDBQueryAllOut    []xq.ChessDBMoveInfo
	chessDBQueryAllErr    error
	chessDBQueryRuleOut   []xq.ChessDBRuleResult
	chessDBQueryRuleErr   error
	chessDBQueueErr       error
	chessDBStoreErr       error
	zobristHashOut        uint64
	zobristHashErr        error
	newImageEncoderErr    error
	closeImageEncoderErr  error
	imageSVGOut           string
	imageSVGErr           error
}

func (f *fakeEngine) ValidateFEN(string) error { return f.validateFENErr }
func (f *fakeEngine) CanonicalFEN(string) (string, error) {
	return f.canonicalFENOut, f.canonicalFENErr
}
func (f *fakeEngine) Turn(string) (xq.Color, error)          { return f.turnOut, f.turnErr }
func (f *fakeEngine) InCheck(string, xq.Color) (bool, error) { return f.inCheck, f.inCheckErr }
func (f *fakeEngine) Status(string) (xq.GameStatus, error)   { return f.statusOut, f.statusErr }
func (f *fakeEngine) Board(string) (xq.Board, error)         { return f.boardOut, f.boardErr }
func (f *fakeEngine) Hash(string) (string, error)            { return f.hashOut, f.hashErr }
func (f *fakeEngine) HalfMoveClock(string) (int, error)      { return f.halfMoveOut, f.halfMoveErr }
func (f *fakeEngine) Rule60(string) (int, error)             { return f.rule60Out, f.rule60Err }
func (f *fakeEngine) Check10(string, xq.Color) (int, error)  { return f.check10Out, f.check10Err }
func (f *fakeEngine) PieceMaterial(string) (xq.PieceMaterial, error) {
	return f.pieceMaterialOut, f.pieceMaterialErr
}
func (f *fakeEngine) BoardMap(string) (map[xq.Square]xq.Piece, error) {
	return f.boardMapOut, f.boardMapErr
}
func (f *fakeEngine) BoardDraw(string) (string, error)   { return f.boardDrawOut, f.boardDrawErr }
func (f *fakeEngine) BoardString(string) (string, error) { return f.boardStringOut, f.boardStringErr }
func (f *fakeEngine) BoardFlip(string, xq.FlipDirection) (xq.Board, error) {
	return f.boardFlipOut, f.boardFlipErr
}
func (f *fakeEngine) PositionString(string) (string, error) {
	return f.positionStringOut, f.positionStringErr
}
func (f *fakeEngine) PositionMarshalText(string) (string, error) {
	return f.positionMarshalTextOut, f.positionMarshalTextErr
}
func (f *fakeEngine) PositionUnmarshalText(string) (string, error) {
	return f.positionUnmarshalTextOut, f.positionUnmarshalTextErr
}
func (f *fakeEngine) PositionMarshalBinary(string) ([]byte, error) {
	return f.positionMarshalBinaryOut, f.positionMarshalBinaryErr
}
func (f *fakeEngine) PositionUnmarshalBinary([]byte) (string, error) {
	return f.positionUnmarshalBinaryOut, f.positionUnmarshalBinaryErr
}
func (f *fakeEngine) ValidMoves(string) ([]xq.Move, error) { return f.validMovesOut, f.validMovesErr }
func (f *fakeEngine) ApplyMove(string, xq.Move) (string, error) {
	return f.applyMoveOut, f.applyMoveErr
}
func (f *fakeEngine) ApplyMoves(string, []xq.Move) (string, error) {
	return f.applyMovesOut, f.applyMovesErr
}
func (f *fakeEngine) DecodeMove(string, string, xq.Notation) (xq.Move, error) {
	return f.decodeMoveOut, f.decodeMoveErr
}
func (f *fakeEngine) EncodeMove(string, xq.Move, xq.Notation) (string, error) {
	return f.encodeMoveOut, f.encodeMoveErr
}
func (f *fakeEngine) ParsePGN(string) (xq.PGNGame, error)  { return f.parsePGNOut, f.parsePGNErr }
func (f *fakeEngine) EncodePGN(xq.PGNGame) (string, error) { return f.encodePGNOut, f.encodePGNErr }
func (f *fakeEngine) NewGame(string, ...xq.GameOption) (xq.GameHandle, error) {
	return f.newGameOut, f.newGameErr
}
func (f *fakeEngine) GameMove(xq.GameHandle, xq.Move) error   { return f.gameMoveErr }
func (f *fakeEngine) GameUndo(xq.GameHandle) (xq.Move, error) { return f.gameUndoOut, f.gameUndoErr }
func (f *fakeEngine) GameUndoN(xq.GameHandle, int) ([]xq.Move, error) {
	return f.gameUndoNOut, f.gameUndoNErr
}
func (f *fakeEngine) GameFEN(xq.GameHandle) (string, error) { return f.gameFENOut, f.gameFENErr }
func (f *fakeEngine) GameMoves(xq.GameHandle) ([]xq.Move, error) {
	return f.gameMovesOut, f.gameMovesErr
}
func (f *fakeEngine) GamePositions(xq.GameHandle) ([]string, error) {
	return f.gamePositionsOut, f.gamePositionsErr
}
func (f *fakeEngine) GameComments(xq.GameHandle) []string { return f.gameCommentsOut }
func (f *fakeEngine) GameStatus(xq.GameHandle) (xq.GameStatus, error) {
	return f.gameStatusOut, f.gameStatusErr
}
func (f *fakeEngine) GamePGN(xq.GameHandle) (string, error) { return f.gamePGNOut, f.gamePGNErr }
func (f *fakeEngine) GameString(xq.GameHandle) (string, error) {
	return f.gameStringOut, f.gameStringErr
}
func (f *fakeEngine) GameMarshalText(xq.GameHandle) (string, error) {
	return f.gameMarshalTextOut, f.gameMarshalTextErr
}
func (f *fakeEngine) GameUnmarshalText(xq.GameHandle, string) error { return f.gameUnmarshalTextErr }
func (f *fakeEngine) GameDraw(xq.GameHandle, xq.DrawReason) error   { return f.gameDrawErr }
func (f *fakeEngine) GameResign(xq.GameHandle, xq.Color) error      { return f.gameResignErr }
func (f *fakeEngine) GameEligibleDraws(xq.GameHandle) ([]xq.DrawReason, error) {
	return f.gameEligibleDrawsOut, f.gameEligibleDrawsErr
}
func (f *fakeEngine) GameClone(xq.GameHandle) (xq.GameHandle, error) {
	return f.gameCloneOut, f.gameCloneErr
}
func (f *fakeEngine) GameSetNotation(xq.GameHandle, xq.Notation) error { return f.gameSetNotationErr }
func (f *fakeEngine) GameTagAdd(xq.GameHandle, string, string) (bool, error) {
	return f.gameTagAddOut, f.gameTagAddErr
}
func (f *fakeEngine) GameTagGet(xq.GameHandle, string) (string, bool, error) {
	return f.gameTagGetOut, f.gameTagGetFound, f.gameTagGetErr
}
func (f *fakeEngine) GameTagRemove(xq.GameHandle, string) (bool, error) {
	return f.gameTagRemoveOut, f.gameTagRemoveErr
}
func (f *fakeEngine) GameJudge(xq.GameHandle) (xq.Verdict, error) {
	return f.gameJudgeOut, f.gameJudgeErr
}
func (f *fakeEngine) GameMoveHistory(xq.GameHandle) ([]xq.MoveHistory, error) {
	return f.gameMoveHistoryOut, f.gameMoveHistoryErr
}
func (f *fakeEngine) CloseGame(xq.GameHandle) error { return f.closeGameErr }
func (f *fakeEngine) OpeningFind([]xq.Move) (*xq.Opening, error) {
	return f.openingFindOut, f.openingFindErr
}
func (f *fakeEngine) OpeningFindExact([]xq.Move) (*xq.Opening, bool, error) {
	return f.openingFindExactOut, f.openingFindExactFound, f.openingFindExactErr
}
func (f *fakeEngine) OpeningPossible([]xq.Move) ([]xq.Opening, error) {
	return f.openingPossibleOut, f.openingPossibleErr
}
func (f *fakeEngine) NewEngine(string) (xq.EngineHandle, error) {
	return f.newEngineOut, f.newEngineErr
}
func (f *fakeEngine) EngineRun(xq.EngineHandle, xq.UCICmd) error { return f.engineRunErr }
func (f *fakeEngine) EngineStop(xq.EngineHandle) error           { return f.engineStopErr }
func (f *fakeEngine) EngineClose(xq.EngineHandle) error          { return f.engineCloseErr }
func (f *fakeEngine) EngineID(xq.EngineHandle) (map[string]string, error) {
	return f.engineIDOut, f.engineIDErr
}
func (f *fakeEngine) EngineOptions(xq.EngineHandle) (map[string]xq.UCIOption, error) {
	return f.engineOptionsOut, f.engineOptionsErr
}
func (f *fakeEngine) EngineSearchResults(xq.EngineHandle) (xq.UCISearchResults, error) {
	return f.engineSearchResultsOut, f.engineSearchResultsErr
}

// === 新增 16 个方法的 fakeEngine stub ===
// Task 9 全量验证时为让 sdk_test.go 编译通过而补齐。

func (f *fakeEngine) EngineKill(xq.EngineHandle) error { return f.engineKillErr }

func (f *fakeEngine) NewChessDBClient(string) (xq.ChessDBClientHandle, error) {
	return 0, f.newChessDBClientErr
}
func (f *fakeEngine) CloseChessDBClient(xq.ChessDBClientHandle) error {
	return f.closeChessDBClientErr
}
func (f *fakeEngine) ChessDBQuery(xq.ChessDBClientHandle, string, xq.ChessDBQueryOptions) ([]xq.ChessDBSuggestedMove, error) {
	return f.chessDBQueryOut, f.chessDBQueryErr
}
func (f *fakeEngine) ChessDBQueryBest(xq.ChessDBClientHandle, string, xq.ChessDBQueryOptions) ([]xq.ChessDBSuggestedMove, error) {
	return f.chessDBQueryBestOut, f.chessDBQueryBestErr
}
func (f *fakeEngine) ChessDBQueryPV(xq.ChessDBClientHandle, string, xq.ChessDBQueryOptions) (xq.ChessDBPV, error) {
	return f.chessDBQueryPVOut, f.chessDBQueryPVErr
}
func (f *fakeEngine) ChessDBQueryScore(xq.ChessDBClientHandle, string, xq.ChessDBQueryOptions) (xq.ChessDBScore, error) {
	return f.chessDBQueryScoreOut, f.chessDBQueryScoreErr
}
func (f *fakeEngine) ChessDBQuerySearch(xq.ChessDBClientHandle, string, xq.ChessDBQueryOptions) ([]xq.ChessDBSuggestedMove, error) {
	return f.chessDBQuerySearchOut, f.chessDBQuerySearchErr
}
func (f *fakeEngine) ChessDBQueryAll(xq.ChessDBClientHandle, string, xq.ChessDBQueryAllOptions) ([]xq.ChessDBMoveInfo, error) {
	return f.chessDBQueryAllOut, f.chessDBQueryAllErr
}
func (f *fakeEngine) ChessDBQueryRule(xq.ChessDBClientHandle, string, []string, xq.ChessDBQueryRuleOptions) ([]xq.ChessDBRuleResult, error) {
	return f.chessDBQueryRuleOut, f.chessDBQueryRuleErr
}
func (f *fakeEngine) ChessDBQueue(xq.ChessDBClientHandle, string) error { return f.chessDBQueueErr }
func (f *fakeEngine) ChessDBStore(xq.ChessDBClientHandle, string, string) error {
	return f.chessDBStoreErr
}

func (f *fakeEngine) ZobristHash(string) (uint64, error) { return f.zobristHashOut, f.zobristHashErr }

func (f *fakeEngine) NewImageEncoder() (xq.ImageEncoderHandle, error) {
	return 0, f.newImageEncoderErr
}
func (f *fakeEngine) CloseImageEncoder(xq.ImageEncoderHandle) error {
	return f.closeImageEncoderErr
}
func (f *fakeEngine) ImageSVG(xq.ImageEncoderHandle, string, ...xq.ImageOption) (string, error) {
	return f.imageSVGOut, f.imageSVGErr
}

func TestSDK_AllForward(t *testing.T) {
	wantMove := xq.Move{S1: xq.Square{File: xq.FileH, Rank: xq.Rank2}, S2: xq.Square{File: xq.FileE, Rank: xq.Rank2}}
	wantMoves := []xq.Move{wantMove}
	eng := &fakeEngine{
		validateFENErr:  errors.New("v"),
		canonicalFENOut: "cfen", canonicalFENErr: errors.New("cf"),
		turnOut: xq.Red, turnErr: errors.New("t"),
		inCheck: true, inCheckErr: errors.New("ic"),
		statusOut: xq.StatusCheck, statusErr: errors.New("s"),
		boardOut: xq.Board{}, boardErr: errors.New("b"),
		hashOut: "abc", hashErr: errors.New("h"),
		halfMoveOut: 1, halfMoveErr: errors.New("hmc"),
		rule60Out: 1, rule60Err: errors.New("r60"),
		check10Out: 1, check10Err: errors.New("c10"),
		pieceMaterialOut: xq.PieceMaterial{"R", "r"}, pieceMaterialErr: errors.New("pm"),
		boardMapOut: map[xq.Square]xq.Piece{}, boardMapErr: errors.New("bm"),
		boardDrawOut: "..", boardDrawErr: errors.New("bd"),
		boardStringOut: "..", boardStringErr: errors.New("bs"),
		boardFlipOut: xq.Board{}, boardFlipErr: errors.New("bf"),
		positionStringOut: "p", positionStringErr: errors.New("ps"),
		positionMarshalTextOut: "p", positionMarshalTextErr: errors.New("pmt"),
		positionUnmarshalTextOut: "p", positionUnmarshalTextErr: errors.New("put"),
		positionMarshalBinaryOut: nil, positionMarshalBinaryErr: errors.New("pmb"),
		positionUnmarshalBinaryOut: "p", positionUnmarshalBinaryErr: errors.New("pub"),
		validMovesOut: wantMoves, validMovesErr: errors.New("vm"),
		applyMoveOut: "newfen", applyMoveErr: errors.New("am"),
		applyMovesOut: "batchfen", applyMovesErr: errors.New("ams"),
		decodeMoveOut: wantMove, decodeMoveErr: errors.New("dm"),
		encodeMoveOut: "h3e3", encodeMoveErr: errors.New("em"),
		parsePGNOut: xq.PGNGame{Result: "*"}, parsePGNErr: errors.New("p"),
		encodePGNOut: "pgn", encodePGNErr: errors.New("ep"),
		newGameOut: 42, newGameErr: errors.New("ng"),
		gameMoveErr: errors.New("gm"),
		gameUndoOut: wantMove, gameUndoErr: errors.New("gu"),
		gameUndoNOut: wantMoves, gameUndoNErr: errors.New("gun"),
		gameFENOut: "fen", gameFENErr: errors.New("gf"),
		gameMovesOut: wantMoves, gameMovesErr: errors.New("gms"),
		gamePositionsOut: []string{"fen"}, gamePositionsErr: errors.New("gps"),
		gameStatusOut: xq.StatusInProgress, gameStatusErr: errors.New("gs"),
		gamePGNOut: "pgn", gamePGNErr: errors.New("gp"),
		gameStringOut: "pgn", gameStringErr: errors.New("gst"),
		gameMarshalTextOut: "pgn", gameMarshalTextErr: errors.New("gmt"),
		gameUnmarshalTextErr: errors.New("gut"),
		gameDrawErr:          errors.New("gd"),
		gameResignErr:        errors.New("gr"),
		gameEligibleDrawsOut: []xq.DrawReason{xq.DrawByAgreement}, gameEligibleDrawsErr: errors.New("ged"),
		gameCloneOut: 99, gameCloneErr: errors.New("gc"),
		gameSetNotationErr: errors.New("gsn"),
		gameTagAddOut:      true, gameTagAddErr: errors.New("gta"),
		gameTagGetOut: "v", gameTagGetFound: true, gameTagGetErr: errors.New("gtg"),
		gameTagRemoveOut: true, gameTagRemoveErr: errors.New("gtr"),
		gameJudgeOut: xq.Verdict{Outcome: xq.OutcomeUnknown}, gameJudgeErr: errors.New("gj"),
		gameMoveHistoryOut: nil, gameMoveHistoryErr: errors.New("gmh"),
		closeGameErr:   errors.New("cg"),
		openingFindOut: &xq.Opening{Code: "B00", Name: "X"}, openingFindErr: errors.New("of"),
		openingFindExactOut: &xq.Opening{Code: "B00", Name: "X"}, openingFindExactFound: true, openingFindExactErr: errors.New("ofe"),
		openingPossibleOut: []xq.Opening{{Code: "C00"}}, openingPossibleErr: errors.New("op"),
		newEngineOut: 77, newEngineErr: errors.New("ne"),
		engineRunErr:   errors.New("er"),
		engineStopErr:  errors.New("es"),
		engineCloseErr: errors.New("ec"),
		engineIDOut:    map[string]string{"name": "x"}, engineIDErr: errors.New("eid"),
		engineOptionsOut: map[string]xq.UCIOption{}, engineOptionsErr: errors.New("eo"),
		engineSearchResultsOut: xq.UCISearchResults{}, engineSearchResultsErr: errors.New("esr"),
	}
	s := &SDK{engine: eng}

	cases := []struct {
		name string
		run  func() error
		want string
	}{
		{"ValidateFEN", func() error { return s.ValidateFEN("x") }, "v"},
		{"CanonicalFEN", func() error { _, err := s.CanonicalFEN("x"); return err }, "cf"},
		{"Turn", func() error { _, err := s.Turn("x"); return err }, "t"},
		{"InCheck", func() error { _, err := s.InCheck("x", xq.Red); return err }, "ic"},
		{"Status", func() error { _, err := s.Status("x"); return err }, "s"},
		{"Board", func() error { _, err := s.Board("x"); return err }, "b"},
		{"Hash", func() error { _, err := s.Hash("x"); return err }, "h"},
		{"HalfMoveClock", func() error { _, err := s.HalfMoveClock("x"); return err }, "hmc"},
		{"Rule60", func() error { _, err := s.Rule60("x"); return err }, "r60"},
		{"Check10", func() error { _, err := s.Check10("x", xq.Red); return err }, "c10"},
		{"PieceMaterial", func() error { _, err := s.PieceMaterial("x"); return err }, "pm"},
		{"BoardMap", func() error { _, err := s.BoardMap("x"); return err }, "bm"},
		{"BoardDraw", func() error { _, err := s.BoardDraw("x"); return err }, "bd"},
		{"BoardString", func() error { _, err := s.BoardString("x"); return err }, "bs"},
		{"BoardFlip", func() error { _, err := s.BoardFlip("x", xq.FlipUpDown); return err }, "bf"},
		{"PositionString", func() error { _, err := s.PositionString("x"); return err }, "ps"},
		{"PositionMarshalText", func() error { _, err := s.PositionMarshalText("x"); return err }, "pmt"},
		{"PositionUnmarshalText", func() error { _, err := s.PositionUnmarshalText("x"); return err }, "put"},
		{"PositionMarshalBinary", func() error { _, err := s.PositionMarshalBinary("x"); return err }, "pmb"},
		{"PositionUnmarshalBinary", func() error { _, err := s.PositionUnmarshalBinary(nil); return err }, "pub"},
		{"ValidMoves", func() error { _, err := s.ValidMoves("x"); return err }, "vm"},
		{"ApplyMove", func() error { _, err := s.ApplyMove("x", wantMove); return err }, "am"},
		{"ApplyMoves", func() error { _, err := s.ApplyMoves("x", wantMoves); return err }, "ams"},
		{"DecodeMove", func() error { _, err := s.DecodeMove("x", "y", xq.NotationUCI); return err }, "dm"},
		{"EncodeMove", func() error { _, err := s.EncodeMove("x", wantMove, xq.NotationUCI); return err }, "em"},
		{"ParsePGN", func() error { _, err := s.ParsePGN("x"); return err }, "p"},
		{"EncodePGN", func() error { _, err := s.EncodePGN(xq.PGNGame{}); return err }, "ep"},
		{"NewGame", func() error { _, err := s.NewGame(""); return err }, "ng"},
		{"GameMove", func() error { return s.GameMove(1, wantMove) }, "gm"},
		{"GameUndo", func() error { _, err := s.GameUndo(1); return err }, "gu"},
		{"GameUndoN", func() error { _, err := s.GameUndoN(1, 1); return err }, "gun"},
		{"GameFEN", func() error { _, err := s.GameFEN(1); return err }, "gf"},
		{"GameMoves", func() error { _, err := s.GameMoves(1); return err }, "gms"},
		{"GamePositions", func() error { _, err := s.GamePositions(1); return err }, "gps"},
		{"GameStatus", func() error { _, err := s.GameStatus(1); return err }, "gs"},
		{"GamePGN", func() error { _, err := s.GamePGN(1); return err }, "gp"},
		{"GameString", func() error { _, err := s.GameString(1); return err }, "gst"},
		{"GameMarshalText", func() error { _, err := s.GameMarshalText(1); return err }, "gmt"},
		{"GameUnmarshalText", func() error { return s.GameUnmarshalText(1, "pgn") }, "gut"},
		{"GameDraw", func() error { return s.GameDraw(1, xq.DrawByAgreement) }, "gd"},
		{"GameResign", func() error { return s.GameResign(1, xq.Red) }, "gr"},
		{"GameEligibleDraws", func() error { _, err := s.GameEligibleDraws(1); return err }, "ged"},
		{"GameClone", func() error { _, err := s.GameClone(1); return err }, "gc"},
		{"GameSetNotation", func() error { return s.GameSetNotation(1, xq.NotationUCI) }, "gsn"},
		{"GameTagAdd", func() error { _, err := s.GameTagAdd(1, "k", "v"); return err }, "gta"},
		{"GameTagGet", func() error { _, _, err := s.GameTagGet(1, "k"); return err }, "gtg"},
		{"GameTagRemove", func() error { _, err := s.GameTagRemove(1, "k"); return err }, "gtr"},
		{"GameJudge", func() error { _, err := s.GameJudge(1); return err }, "gj"},
		{"GameMoveHistory", func() error { _, err := s.GameMoveHistory(1); return err }, "gmh"},
		{"CloseGame", func() error { return s.CloseGame(1) }, "cg"},
		{"OpeningFind", func() error { _, err := s.OpeningFind(nil); return err }, "of"},
		{"OpeningFindExact", func() error { _, _, err := s.OpeningFindExact(nil); return err }, "ofe"},
		{"OpeningPossible", func() error { _, err := s.OpeningPossible(nil); return err }, "op"},
		{"NewEngine", func() error { _, err := s.NewEngine("/path"); return err }, "ne"},
		{"EngineRun", func() error { return s.EngineRun(1, xq.UCICmdIsReady{}) }, "er"},
		{"EngineStop", func() error { return s.EngineStop(1) }, "es"},
		{"EngineClose", func() error { return s.EngineClose(1) }, "ec"},
		{"EngineID", func() error { _, err := s.EngineID(1); return err }, "eid"},
		{"EngineOptions", func() error { _, err := s.EngineOptions(1); return err }, "eo"},
		{"EngineSearchResults", func() error { _, err := s.EngineSearchResults(1); return err }, "esr"},
	}
	for _, c := range cases {
		err := c.run()
		if err == nil || err.Error() != c.want {
			t.Errorf("%s: err = %v, want %s", c.name, err, c.want)
		}
	}
}
