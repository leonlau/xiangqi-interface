package main

import (
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/leonlau/xiangqi"
	xq "github.com/leonlau/xiangqi-interface"
	"github.com/leonlau/xiangqi/opening"
)

type XiangqiEngineImpl struct{}

func NewXiangqiEngine() xq.XiangqiEngine { return &XiangqiEngineImpl{} }

var _ xq.XiangqiEngine = (*XiangqiEngineImpl)(nil)

// === 类型转换 helpers ===

func colorToXiangqi(c xq.Color) xiangqi.Color {
	switch c {
	case xq.Red:
		return xiangqi.Red
	case xq.Black:
		return xiangqi.Black
	}
	return xiangqi.NoColor
}

func colorToXq(c xiangqi.Color) xq.Color {
	switch c {
	case xiangqi.Red:
		return xq.Red
	case xiangqi.Black:
		return xq.Black
	}
	return xq.NoColor
}

func pieceTypeToXiangqi(t xq.PieceType) xiangqi.PieceType {
	switch t {
	case xq.King:
		return xiangqi.King
	case xq.Advisor:
		return xiangqi.Advisor
	case xq.Elephant:
		return xiangqi.Bishop
	case xq.Horse:
		return xiangqi.Knight
	case xq.Rook:
		return xiangqi.Rook
	case xq.Cannon:
		return xiangqi.Cannon
	case xq.Pawn:
		return xiangqi.Pawn
	}
	return xiangqi.NoPieceType
}

func pieceTypeToXq(t xiangqi.PieceType) xq.PieceType {
	switch t {
	case xiangqi.King:
		return xq.King
	case xiangqi.Advisor:
		return xq.Advisor
	case xiangqi.Bishop:
		return xq.Elephant
	case xiangqi.Knight:
		return xq.Horse
	case xiangqi.Rook:
		return xq.Rook
	case xiangqi.Cannon:
		return xq.Cannon
	case xiangqi.Pawn:
		return xq.Pawn
	}
	return xq.NoPieceType
}

type gameState struct {
	game     *xiangqi.Game
	startFEN string
	opts     []xq.GameOption
}

var (
	gameMu      sync.Mutex
	gameCounter atomic.Int64
	games       = map[xq.GameHandle]*gameState{}

	bookMu sync.Mutex
	book   *opening.BookECO
)

func getBook() *opening.BookECO {
	bookMu.Lock()
	defer bookMu.Unlock()
	if book == nil {
		book = opening.NewBookECO()
	}
	return book
}

func getGame(h xq.GameHandle) (*gameState, error) {
	gameMu.Lock()
	defer gameMu.Unlock()
	gs, ok := games[h]
	if !ok {
		return nil, fmt.Errorf("invalid game handle %d", h)
	}
	return gs, nil
}

// === FEN ===

func (XiangqiEngineImpl) ValidateFEN(fen string) error { _, err := parsePosition(fen); return err }

func (XiangqiEngineImpl) CanonicalFEN(fen string) (string, error) {
	pos, err := parsePosition(fen)
	if err != nil {
		return "", err
	}
	data, err := pos.MarshalText()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// === Position queries ===

func (XiangqiEngineImpl) Turn(fen string) (xq.Color, error) {
	pos, err := parsePosition(fen)
	if err != nil {
		return xq.NoColor, err
	}
	return colorToXq(pos.Turn()), nil
}

func (XiangqiEngineImpl) InCheck(fen string, c xq.Color) (bool, error) {
	pos, err := parsePosition(fen)
	if err != nil {
		return false, err
	}
	xc := colorToXiangqi(c)
	board := pos.Board()
	var kingSq xiangqi.Square
	found := false
	for r := xiangqi.Rank(0); r <= xiangqi.Rank(9); r++ {
		for f := xiangqi.File(0); f <= xiangqi.File(8); f++ {
			sq := xiangqi.NewSquare(f, r)
			p := board.Piece(sq)
			if p.Type() == xiangqi.King && p.Color() == xc {
				kingSq = sq
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		return false, fmt.Errorf("king of color %v not found", c)
	}
	var opp []*xiangqi.Move
	if pos.Turn() == xc {
		data, err := pos.MarshalText()
		if err != nil {
			return false, err
		}
		parts := strings.Split(string(data), " ")
		if parts[1] == "w" {
			parts[1] = "b"
		} else {
			parts[1] = "w"
		}
		flipped, err := parsePosition(strings.Join(parts, " "))
		if err != nil {
			return false, err
		}
		opp = flipped.ValidMoves()
	} else {
		opp = pos.ValidMoves()
	}
	for _, m := range opp {
		if m.S2() == kingSq {
			return true, nil
		}
	}
	return false, nil
}

func (XiangqiEngineImpl) Status(fen string) (xq.GameStatus, error) {
	pos, err := parsePosition(fen)
	if err != nil {
		return xq.StatusInProgress, err
	}
	method := pos.Status()
	switch method {
	case xiangqi.Checkmate:
		return xq.StatusCheckmate, nil
	case xiangqi.Stalemate:
		return xq.StatusStalemate, nil
	case xiangqi.ThreefoldRepetition, xiangqi.FivefoldRepetition,
		xiangqi.FiftyMoveRule, xiangqi.SeventyFiveMoveRule, xiangqi.DrawOffer:
		return xq.StatusDraw, nil
	}
	inCheck, _ := XiangqiEngineImpl{}.InCheck(fen, colorToXq(pos.Turn()))
	if inCheck {
		return xq.StatusCheck, nil
	}
	return xq.StatusInProgress, nil
}

func (XiangqiEngineImpl) Board(fen string) (xq.Board, error) {
	pos, err := parsePosition(fen)
	if err != nil {
		return xq.Board{}, err
	}
	board := pos.Board()
	var out xq.Board
	for r := xq.Rank0; r <= xq.Rank9; r++ {
		for f := xq.FileA; f <= xq.FileI; f++ {
			sq := xiangqi.NewSquare(xiangqi.File(f), xiangqi.Rank(r))
			p := board.Piece(sq)
			out[int(r)*9+int(f)] = xq.Piece{
				Type:  pieceTypeToXq(p.Type()),
				Color: colorToXq(p.Color()),
			}
		}
	}
	return out, nil
}

func (XiangqiEngineImpl) Hash(fen string) (string, error) {
	pos, err := parsePosition(fen)
	if err != nil {
		return "", err
	}
	h := pos.Hash()
	return hex.EncodeToString(h[:]), nil
}

func (XiangqiEngineImpl) HalfMoveClock(fen string) (int, error) {
	pos, err := parsePosition(fen)
	if err != nil {
		return 0, err
	}
	return pos.HalfMoveClock(), nil
}

func (XiangqiEngineImpl) Rule60(fen string) (int, error) {
	pos, err := parsePosition(fen)
	if err != nil {
		return 0, err
	}
	return pos.Rule60(), nil
}

func (XiangqiEngineImpl) Check10(fen string, c xq.Color) (int, error) {
	pos, err := parsePosition(fen)
	if err != nil {
		return 0, err
	}
	return pos.Check10(colorToXiangqi(c)), nil
}

func (XiangqiEngineImpl) PieceMaterial(fen string) (xq.PieceMaterial, error) {
	pos, err := parsePosition(fen)
	if err != nil {
		return xq.PieceMaterial{}, err
	}
	pm := pos.Board().PieceMaterial()
	return xq.PieceMaterial{pm[0], pm[1]}, nil
}

func (XiangqiEngineImpl) BoardMap(fen string) (map[xq.Square]xq.Piece, error) {
	board, err := XiangqiEngineImpl{}.Board(fen)
	if err != nil {
		return nil, err
	}
	return board.AsMap(), nil
}

func (XiangqiEngineImpl) BoardDraw(fen string) (string, error) {
	pos, err := parsePosition(fen)
	if err != nil {
		return "", err
	}
	return pos.Board().Draw(), nil
}

// === Move generation & state ===

func (XiangqiEngineImpl) ValidMoves(fen string) ([]xq.Move, error) {
	pos, err := parsePosition(fen)
	if err != nil {
		return nil, err
	}
	internal := pos.ValidMoves()
	out := make([]xq.Move, len(internal))
	for i, m := range internal {
		out[i] = convertMove(m)
	}
	return out, nil
}

func (XiangqiEngineImpl) ApplyMove(fen string, m xq.Move) (string, error) {
	pos, err := parsePosition(fen)
	if err != nil {
		return "", err
	}
	uci := formatUCI(m.S1) + formatUCI(m.S2)
	internal, err := xiangqi.UCINotation{}.Decode(pos, uci)
	if err != nil {
		return "", fmt.Errorf("invalid move %s: %w", uci, err)
	}
	updated := pos.Update(internal)
	data, err := updated.MarshalText()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (XiangqiEngineImpl) ApplyMoves(fen string, moves []xq.Move) (string, error) {
	current := fen
	for _, m := range moves {
		next, err := XiangqiEngineImpl{}.ApplyMove(current, m)
		if err != nil {
			return "", err
		}
		current = next
	}
	return current, nil
}

// === Notation (multi-format) ===

func pickNotation(n xq.Notation) (xiangqi.Notation, error) {
	switch n {
	case xq.NotationUCI:
		return xiangqi.UCINotation{}, nil
	case xq.NotationChinese:
		return xiangqi.ChineseCharNotation{}, nil
	case xq.NotationChessDB:
		return xiangqi.ChessDBNotation{}, nil
	case xq.NotationFEN:
		return xiangqi.FENNotation{}, nil
	case xq.NotationUCCI:
		return xiangqi.UCCINotation{}, nil
	}
	return nil, fmt.Errorf("unknown notation %v", n)
}

func (XiangqiEngineImpl) DecodeMove(fen, notation string, n xq.Notation) (xq.Move, error) {
	pos, err := parsePosition(fen)
	if err != nil {
		return xq.Move{}, err
	}
	not, err := pickNotation(n)
	if err != nil {
		return xq.Move{}, err
	}
	m, err := not.Decode(pos, notation)
	if err != nil {
		return xq.Move{}, fmt.Errorf("decode %q with %v: %w", notation, n, err)
	}
	return convertMove(m), nil
}

func (XiangqiEngineImpl) EncodeMove(fen string, m xq.Move, n xq.Notation) (string, error) {
	pos, err := parsePosition(fen)
	if err != nil {
		return "", err
	}
	not, err := pickNotation(n)
	if err != nil {
		return "", err
	}
	if _, isUCI := not.(xiangqi.UCINotation); isUCI {
		return formatUCI(m.S1) + formatUCI(m.S2), nil
	}
	uci := formatUCI(m.S1) + formatUCI(m.S2)
	internal, err := xiangqi.UCINotation{}.Decode(pos, uci)
	if err != nil {
		return "", fmt.Errorf("invalid move %s: %w", uci, err)
	}
	return not.Encode(pos, internal), nil
}

// === PGN ===

func (XiangqiEngineImpl) ParsePGN(pgn string) (xq.PGNGame, error) {
	opt, err := xiangqi.PGN(strings.NewReader(pgn))
	if err != nil {
		return xq.PGNGame{}, fmt.Errorf("parse pgn: %w", err)
	}
	g := xiangqi.NewGame(opt)
	moves := g.Moves()
	comments := g.Comments()
	out := xq.PGNGame{
		Tags:     make([]xq.PGNTag, 0, len(g.TagPairs())),
		Moves:    make([]xq.Move, 0, len(moves)),
		Comments: flattenComments(comments),
		Result:   string(g.Outcome()),
	}
	for _, t := range g.TagPairs() {
		out.Tags = append(out.Tags, xq.PGNTag{Name: t.Key, Value: t.Value})
	}
	for _, m := range moves {
		out.Moves = append(out.Moves, convertMove(m))
	}
	return out, nil
}

func (XiangqiEngineImpl) EncodePGN(game xq.PGNGame) (string, error) {
	tags := make([]*xiangqi.TagPair, 0, len(game.Tags))
	for _, t := range game.Tags {
		tags = append(tags, &xiangqi.TagPair{Key: t.Name, Value: t.Value})
	}
	g := xiangqi.NewGame(
		xiangqi.UseNotation(xiangqi.UCINotation{}),
		xiangqi.TagPairs(tags),
	)
	for _, m := range game.Moves {
		uci := formatUCI(m.S1) + formatUCI(m.S2)
		if err := g.MoveStr(uci); err != nil {
			return "", fmt.Errorf("encode pgn at move %s: %w", uci, err)
		}
	}
	return g.String(), nil
}

// === Game (stateful) ===

// applyGameOptions 把 GameOption 翻译成 xiangqi.NewGame 的 option 列表。
func applyGameOptions(opts []xq.GameOption) ([]func(*xiangqi.Game), error) {
	var out []func(*xiangqi.Game)
	for _, o := range opts {
		switch v := o.(type) {
		case xq.UseRulesOpt:
			// 接口层保留 "asian" 别名(历史惯例),实际上游仅识别
			// "Pikafish-2023";其他名称按 no-op 处理。
			ruleset := v.Rules
			if ruleset == "asian" {
				ruleset = "Pikafish-2023"
			}
			out = append(out, xiangqi.UseRules(ruleset))
		case xq.IgnoreRulesOpt:
			out = append(out, xiangqi.IgnoreRules())
		case xq.UseNotationOpt:
			n, err := pickNotation(v.N)
			if err != nil {
				return nil, err
			}
			out = append(out, xiangqi.UseNotation(n))
		default:
			return nil, fmt.Errorf("unknown GameOption %T", o)
		}
	}
	return out, nil
}

func (XiangqiEngineImpl) NewGame(startFEN string, opts ...xq.GameOption) (xq.GameHandle, error) {
	if startFEN == "" {
		startFEN = "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"
	}
	xiangqiOpts, err := applyGameOptions(opts)
	if err != nil {
		return xq.InvalidGameHandle, err
	}
	opt, err := xiangqi.FEN(startFEN)
	if err != nil {
		return xq.InvalidGameHandle, err
	}
	xiangqiOpts = append(xiangqiOpts, opt)
	g := xiangqi.NewGame(xiangqiOpts...)
	h := xq.GameHandle(gameCounter.Add(1))
	gameMu.Lock()
	games[h] = &gameState{game: g, startFEN: startFEN, opts: opts}
	gameMu.Unlock()
	return h, nil
}

// rebuildGame 从 startFEN 重建 xiangqi.Game 并应用 opts(用于 undo)。
func rebuildGame(startFEN string, opts []xq.GameOption) (*xiangqi.Game, error) {
	xiangqiOpts, err := applyGameOptions(opts)
	if err != nil {
		return nil, err
	}
	opt, err := xiangqi.FEN(startFEN)
	if err != nil {
		return nil, err
	}
	xiangqiOpts = append(xiangqiOpts, opt)
	return xiangqi.NewGame(xiangqiOpts...), nil
}

func (XiangqiEngineImpl) GameMove(h xq.GameHandle, m xq.Move) error {
	gs, err := getGame(h)
	if err != nil {
		return err
	}
	pos := gs.game.Position()
	uci := formatUCI(m.S1) + formatUCI(m.S2)
	internal, err := xiangqi.UCINotation{}.Decode(pos, uci)
	if err != nil {
		return fmt.Errorf("invalid move %s: %w", uci, err)
	}
	return gs.game.Move(internal)
}

func (XiangqiEngineImpl) GameUndo(h xq.GameHandle) (xq.Move, error) {
	gs, err := getGame(h)
	if err != nil {
		return xq.Move{}, err
	}
	moves := gs.game.Moves()
	if len(moves) == 0 {
		return xq.Move{}, fmt.Errorf("no moves to undo")
	}
	last := moves[len(moves)-1]
	lastConverted := convertMove(last)
	// 重建对局:从 startFEN 重放前 n-1 步
	prev := moves[:len(moves)-1]
	newG, err := rebuildGame(gs.startFEN, gs.opts)
	if err != nil {
		return xq.Move{}, err
	}
	for _, m := range prev {
		if err := newG.Move(m); err != nil {
			return xq.Move{}, err
		}
	}
	gameMu.Lock()
	gs.game = newG
	gameMu.Unlock()
	return lastConverted, nil
}

func (XiangqiEngineImpl) GameUndoN(h xq.GameHandle, n int) ([]xq.Move, error) {
	if n <= 0 {
		return nil, nil
	}
	undone := make([]xq.Move, 0, n)
	for i := 0; i < n; i++ {
		m, err := XiangqiEngineImpl{}.GameUndo(h)
		if err != nil {
			return undone, err
		}
		undone = append(undone, m)
	}
	return undone, nil
}

func (XiangqiEngineImpl) GameFEN(h xq.GameHandle) (string, error) {
	gs, err := getGame(h)
	if err != nil {
		return "", err
	}
	return gs.game.FEN(), nil
}

func (XiangqiEngineImpl) GameMoves(h xq.GameHandle) ([]xq.Move, error) {
	gs, err := getGame(h)
	if err != nil {
		return nil, err
	}
	internal := gs.game.Moves()
	out := make([]xq.Move, len(internal))
	for i, m := range internal {
		out[i] = convertMove(m)
	}
	return out, nil
}

func (XiangqiEngineImpl) GameComments(h xq.GameHandle) []string {
	gs, err := getGame(h)
	if err != nil {
		return nil
	}
	return flattenComments(gs.game.Comments())
}

func (XiangqiEngineImpl) GameStatus(h xq.GameHandle) (xq.GameStatus, error) {
	gs, err := getGame(h)
	if err != nil {
		return xq.StatusInProgress, err
	}
	method := gs.game.Method()
	switch method {
	case xiangqi.Checkmate:
		return xq.StatusCheckmate, nil
	case xiangqi.Stalemate:
		return xq.StatusStalemate, nil
	case xiangqi.ThreefoldRepetition, xiangqi.FivefoldRepetition,
		xiangqi.FiftyMoveRule, xiangqi.SeventyFiveMoveRule, xiangqi.DrawOffer:
		return xq.StatusDraw, nil
	}
	fen := gs.game.FEN()
	inCheck, _ := XiangqiEngineImpl{}.InCheck(fen, colorToXq(gs.game.Position().Turn()))
	if inCheck {
		return xq.StatusCheck, nil
	}
	return xq.StatusInProgress, nil
}

func (XiangqiEngineImpl) GamePGN(h xq.GameHandle) (string, error) {
	gs, err := getGame(h)
	if err != nil {
		return "", err
	}
	return gs.game.String(), nil
}

func (XiangqiEngineImpl) GameDraw(h xq.GameHandle, reason xq.DrawReason) error {
	gs, err := getGame(h)
	if err != nil {
		return err
	}
	var method xiangqi.Method
	switch reason {
	case xq.DrawByAgreement:
		method = xiangqi.DrawOffer
	case xq.DrawBy50MoveRule:
		method = xiangqi.FiftyMoveRule
	case xq.DrawByRepetition:
		method = xiangqi.ThreefoldRepetition
	case xq.DrawByStalemate:
		method = xiangqi.Stalemate
	case xq.DrawByInsufficientMaterial:
		// 子力不足由裁判自动裁决,不支持手动宣告和棋。
		return fmt.Errorf("draw by insufficient material is adjudicated automatically, not a manual draw method")
	default:
		return fmt.Errorf("unknown draw reason %v", reason)
	}
	return gs.game.Draw(method)
}

func (XiangqiEngineImpl) GameResign(h xq.GameHandle, c xq.Color) error {
	gs, err := getGame(h)
	if err != nil {
		return err
	}
	gs.game.Resign(colorToXiangqi(c))
	return nil
}

func (XiangqiEngineImpl) CloseGame(h xq.GameHandle) error {
	gameMu.Lock()
	defer gameMu.Unlock()
	if _, ok := games[h]; !ok {
		return fmt.Errorf("invalid game handle %d", h)
	}
	delete(games, h)
	return nil
}

// === Opening book ===

func (XiangqiEngineImpl) OpeningFind(moves []xq.Move) (*xq.Opening, error) {
	opt, err := xiangqi.FEN("rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1")
	if err != nil {
		return nil, err
	}
	g := xiangqi.NewGame(opt)
	internal := make([]*xiangqi.Move, 0, len(moves))
	for _, m := range moves {
		pos := g.Position()
		uci := formatUCI(m.S1) + formatUCI(m.S2)
		im, err := xiangqi.UCINotation{}.Decode(pos, uci)
		if err != nil {
			return nil, fmt.Errorf("invalid move %s: %w", uci, err)
		}
		if err := g.Move(im); err != nil {
			return nil, err
		}
		internal = append(internal, im)
	}
	o := getBook().Find(internal)
	if o == nil {
		return nil, nil
	}
	return &xq.Opening{
		Code:      o.Code(),
		Name:      o.Title(),
		Variation: o.Moves(),
	}, nil
}

func (XiangqiEngineImpl) OpeningPossible(moves []xq.Move) ([]xq.Opening, error) {
	opt, err := xiangqi.FEN("rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1")
	if err != nil {
		return nil, err
	}
	g := xiangqi.NewGame(opt)
	internal := make([]*xiangqi.Move, 0, len(moves))
	for _, m := range moves {
		pos := g.Position()
		uci := formatUCI(m.S1) + formatUCI(m.S2)
		im, err := xiangqi.UCINotation{}.Decode(pos, uci)
		if err != nil {
			return nil, fmt.Errorf("invalid move %s: %w", uci, err)
		}
		if err := g.Move(im); err != nil {
			return nil, err
		}
		internal = append(internal, im)
	}
	os := getBook().Possible(internal)
	out := make([]xq.Opening, 0, len(os))
	for _, o := range os {
		out = append(out, xq.Opening{
			Code:      o.Code(),
			Name:      o.Title(),
			Variation: o.Moves(),
		})
	}
	return out, nil
}

// === Position serialize/flip ===

func (XiangqiEngineImpl) PositionString(fen string) (string, error) {
	pos, err := parsePosition(fen)
	if err != nil {
		return "", err
	}
	return pos.String(), nil
}

func (XiangqiEngineImpl) BoardString(fen string) (string, error) {
	pos, err := parsePosition(fen)
	if err != nil {
		return "", err
	}
	return pos.Board().String(), nil
}

func (XiangqiEngineImpl) BoardFlip(fen string, dir xq.FlipDirection) (xq.Board, error) {
	pos, err := parsePosition(fen)
	if err != nil {
		return xq.Board{}, err
	}
	var xd xiangqi.FlipDirection
	switch dir {
	case xq.FlipUpDown:
		xd = xiangqi.UpDown
	case xq.FlipLeftRight:
		xd = xiangqi.LeftRight
	default:
		return xq.Board{}, fmt.Errorf("unknown flip direction %v", dir)
	}
	flipped := pos.Board().Flip(xd)
	var out xq.Board
	for r := xq.Rank0; r <= xq.Rank9; r++ {
		for f := xq.FileA; f <= xq.FileI; f++ {
			sq := xiangqi.NewSquare(xiangqi.File(f), xiangqi.Rank(r))
			p := flipped.Piece(sq)
			out[int(r)*9+int(f)] = xq.Piece{
				Type:  pieceTypeToXq(p.Type()),
				Color: colorToXq(p.Color()),
			}
		}
	}
	return out, nil
}

func (XiangqiEngineImpl) PositionMarshalText(fen string) (string, error) {
	pos, err := parsePosition(fen)
	if err != nil {
		return "", err
	}
	data, err := pos.MarshalText()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (XiangqiEngineImpl) PositionUnmarshalText(text string) (string, error) {
	var pos xiangqi.Position
	if err := pos.UnmarshalText([]byte(text)); err != nil {
		return "", err
	}
	data, err := pos.MarshalText()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (XiangqiEngineImpl) PositionMarshalBinary(fen string) ([]byte, error) {
	pos, err := parsePosition(fen)
	if err != nil {
		return nil, err
	}
	return pos.MarshalBinary()
}

func (XiangqiEngineImpl) PositionUnmarshalBinary(data []byte) (string, error) {
	var pos xiangqi.Position
	if err := pos.UnmarshalBinary(data); err != nil {
		return "", err
	}
	out, err := pos.MarshalText()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// === Game extended ===

func (XiangqiEngineImpl) GamePositions(h xq.GameHandle) ([]string, error) {
	gs, err := getGame(h)
	if err != nil {
		return nil, err
	}
	positions := gs.game.Positions()
	out := make([]string, 0, len(positions))
	for _, p := range positions {
		data, err := p.MarshalText()
		if err != nil {
			return nil, err
		}
		out = append(out, string(data))
	}
	return out, nil
}

func (XiangqiEngineImpl) GameString(h xq.GameHandle) (string, error) {
	return XiangqiEngineImpl{}.GamePGN(h)
}

func (XiangqiEngineImpl) GameMarshalText(h xq.GameHandle) (string, error) {
	gs, err := getGame(h)
	if err != nil {
		return "", err
	}
	data, err := gs.game.MarshalText()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (XiangqiEngineImpl) GameUnmarshalText(h xq.GameHandle, pgn string) error {
	gs, err := getGame(h)
	if err != nil {
		return err
	}
	return gs.game.UnmarshalText([]byte(pgn))
}

func (XiangqiEngineImpl) GameEligibleDraws(h xq.GameHandle) ([]xq.DrawReason, error) {
	gs, err := getGame(h)
	if err != nil {
		return nil, err
	}
	methods := gs.game.EligibleDraws()
	out := make([]xq.DrawReason, 0, len(methods))
	for _, m := range methods {
		switch m {
		case xiangqi.DrawOffer:
			out = append(out, xq.DrawByAgreement)
		case xiangqi.FiftyMoveRule, xiangqi.SeventyFiveMoveRule:
			out = append(out, xq.DrawBy50MoveRule)
		case xiangqi.ThreefoldRepetition, xiangqi.FivefoldRepetition:
			out = append(out, xq.DrawByRepetition)
		case xiangqi.Stalemate:
			out = append(out, xq.DrawByStalemate)
		}
	}
	return out, nil
}

func (XiangqiEngineImpl) GameClone(h xq.GameHandle) (xq.GameHandle, error) {
	src, err := getGame(h)
	if err != nil {
		return xq.InvalidGameHandle, err
	}
	cloned := src.game.Clone()
	newH := xq.GameHandle(gameCounter.Add(1))
	gameMu.Lock()
	games[newH] = &gameState{
		game:     cloned,
		startFEN: src.startFEN,
		opts:     src.opts,
	}
	gameMu.Unlock()
	return newH, nil
}

func (XiangqiEngineImpl) GameSetNotation(h xq.GameHandle, n xq.Notation) error {
	gs, err := getGame(h)
	if err != nil {
		return err
	}
	not, err := pickNotation(n)
	if err != nil {
		return err
	}
	gs.game.SetNotation(not)
	return nil
}

func (XiangqiEngineImpl) GameTagAdd(h xq.GameHandle, name, value string) (bool, error) {
	gs, err := getGame(h)
	if err != nil {
		return false, err
	}
	return gs.game.AddTagPair(name, value), nil
}

func (XiangqiEngineImpl) GameTagGet(h xq.GameHandle, name string) (string, bool, error) {
	gs, err := getGame(h)
	if err != nil {
		return "", false, err
	}
	tp := gs.game.GetTagPair(name)
	if tp == nil {
		return "", false, nil
	}
	return tp.Value, true, nil
}

func (XiangqiEngineImpl) GameTagRemove(h xq.GameHandle, name string) (bool, error) {
	gs, err := getGame(h)
	if err != nil {
		return false, err
	}
	return gs.game.RemoveTagPair(name), nil
}

// === Opening book extended ===

func (XiangqiEngineImpl) OpeningFindExact(moves []xq.Move) (*xq.Opening, bool, error) {
	opt, err := xiangqi.FEN("rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1")
	if err != nil {
		return nil, false, err
	}
	g := xiangqi.NewGame(opt)
	internal := make([]*xiangqi.Move, 0, len(moves))
	for _, m := range moves {
		pos := g.Position()
		uci := formatUCI(m.S1) + formatUCI(m.S2)
		im, err := xiangqi.UCINotation{}.Decode(pos, uci)
		if err != nil {
			return nil, false, fmt.Errorf("invalid move %s: %w", uci, err)
		}
		if err := g.Move(im); err != nil {
			return nil, false, err
		}
		internal = append(internal, im)
	}
	o, exact := getBook().FindWithExact(internal)
	if o == nil {
		return nil, false, nil
	}
	return &xq.Opening{
		Code:      o.Code(),
		Name:      o.Title(),
		Variation: o.Moves(),
	}, exact, nil
}

// === Judge & MoveHistory ===

func outcomeToXq(o xiangqi.Outcome) xq.Outcome {
	switch o {
	case xiangqi.WhiteWon:
		return xq.OutcomeRedWon
	case xiangqi.BlackWon:
		return xq.OutcomeBlackWon
	case xiangqi.Draw:
		return xq.OutcomeDraw
	}
	return xq.OutcomeUnknown
}

func methodToXq(m xiangqi.Method) xq.Method {
	switch m {
	case xiangqi.Checkmate:
		return xq.MethodCheckmate
	case xiangqi.Resignation:
		return xq.MethodResignation
	case xiangqi.DrawOffer:
		return xq.MethodDrawOffer
	case xiangqi.Stalemate:
		return xq.MethodStalemate
	case xiangqi.ThreefoldRepetition:
		return xq.MethodThreefoldRepetition
	case xiangqi.FivefoldRepetition:
		return xq.MethodFivefoldRepetition
	case xiangqi.FiftyMoveRule:
		return xq.MethodFiftyMoveRule
	case xiangqi.SeventyFiveMoveRule:
		return xq.MethodSeventyFiveMoveRule
	case xiangqi.InsufficientMaterial:
		return xq.MethodInsufficientMaterial
	case xiangqi.PerpetualCheck:
		return xq.MethodPerpetualCheck
	case xiangqi.PerpetualChase:
		return xq.MethodPerpetualChase
	case xiangqi.Repetition:
		return xq.MethodRepetition
	case xiangqi.Rule60:
		return xq.MethodRule60
	}
	return xq.MethodNone
}

func (XiangqiEngineImpl) GameJudge(h xq.GameHandle) (xq.Verdict, error) {
	gs, err := getGame(h)
	if err != nil {
		return xq.Verdict{}, err
	}
	// 收集 Position 历史和 Move 历史
	positions := gs.game.Positions()
	moves := gs.game.Moves()
	posSlice := make([]*xiangqi.Position, len(positions))
	for i, p := range positions {
		posSlice[i] = p
	}
	moveSlice := make([]*xiangqi.Move, len(moves))
	for i, m := range moves {
		moveSlice[i] = m
	}
	curPos := gs.game.Position()
	verdict := gs.game.Judge().Update(curPos, posSlice, moveSlice)
	return xq.Verdict{
		Outcome: outcomeToXq(verdict.Outcome),
		Method:  methodToXq(verdict.Method),
		Reason:  verdict.Reason,
	}, nil
}

func (XiangqiEngineImpl) GameMoveHistory(h xq.GameHandle) ([]xq.MoveHistory, error) {
	gs, err := getGame(h)
	if err != nil {
		return nil, err
	}
	internal := gs.game.MoveHistory()
	out := make([]xq.MoveHistory, 0, len(internal))
	for _, mh := range internal {
		pre, _ := mh.PrePosition.MarshalText()
		post, _ := mh.PostPosition.MarshalText()
		out = append(out, xq.MoveHistory{
			PreFEN:   string(pre),
			PostFEN:  string(post),
			Move:     convertMove(mh.Move),
			Comments: append([]string(nil), mh.Comments...),
		})
	}
	return out, nil
}
