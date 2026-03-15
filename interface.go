package xiangqiinterface

import "time"

type Color int8

const (
	NoColor Color = iota
	Red
	Black
)

func (c Color) String() string {
	switch c {
	case Red:
		return "w"
	case Black:
		return "b"
	}
	return "-"
}

// Other 返回对手颜色。
func (c Color) Other() Color {
	if c == Red {
		return Black
	}
	return Red
}

// Name 返回颜色全名(中文/英文)。
func (c Color) Name() string {
	switch c {
	case Red:
		return "red"
	case Black:
		return "black"
	}
	return "none"
}

type File int8

const (
	FileA File = iota
	FileB
	FileC
	FileD
	FileE
	FileF
	FileG
	FileH
	FileI
)

func (f File) String() string { return string(rune('a' + int(f))) }

// Flip 返回镜像列(从对方视角看)。a↔i, b↔h, c↔g, d↔f, e 保持。
func (f File) Flip() File { return File(8 - int(f)) }

// FlipSquare 翻转 Square 的列。
func (s Square) FlipSquare() Square {
	return Square{File: s.File.Flip(), Rank: s.Rank}
}

type Rank int8

const (
	Rank0 Rank = iota
	Rank1
	Rank2
	Rank3
	Rank4
	Rank5
	Rank6
	Rank7
	Rank8
	Rank9
)

func (r Rank) String() string { return string(rune('0' + int(r))) }

type PieceType int8

const (
	NoPieceType PieceType = iota
	King
	Advisor
	Elephant
	Horse
	Rook
	Cannon
	Pawn
)

// AllPieceTypes 返回全部 7 种棋子类型(不含 NoPieceType)。
func AllPieceTypes() [7]PieceType {
	return [7]PieceType{King, Advisor, Elephant, Horse, Rook, Cannon, Pawn}
}

func (t PieceType) String() string {
	switch t {
	case King:
		return "K"
	case Advisor:
		return "A"
	case Elephant:
		return "E"
	case Horse:
		return "H"
	case Rook:
		return "R"
	case Cannon:
		return "C"
	case Pawn:
		return "P"
	}
	return "."
}

type Square struct {
	File File
	Rank Rank
}

func (s Square) String() string {
	return s.File.String() + s.Rank.String()
}

// NewSquare 构造一个 Square。
func NewSquare(f File, r Rank) Square { return Square{File: f, Rank: r} }

// FileFlip 返回列镜像后的 Square。
func (s Square) FileFlip() Square { return s.FlipSquare() }

type Piece struct {
	Type  PieceType
	Color Color
}

// NewPiece 构造一个 Piece。
func NewPiece(t PieceType, c Color) Piece { return Piece{Type: t, Color: c} }

func (p Piece) String() string {
	if p.Type == NoPieceType {
		return "."
	}
	var c byte = 'w'
	if p.Color == Black {
		c = 'b'
	}
	return string(c) + p.Type.String()
}

// MoveTag 是走子的位掩码标记。
type MoveTag uint16

const (
	TagCapture MoveTag = 1 << iota
	TagCheck
	TagInCheck // 走子后己方被将军(非法)
	TagChase
	TagTrade
)

// Board 表示 9 列 × 10 行的棋盘。索引 = int(Rank)*9 + int(File)。
type Board [90]Piece

// PieceAt 返回指定格子的棋子(空位返回 NoPiece)。
func (b Board) PieceAt(s Square) Piece {
	return b[int(s.Rank)*9+int(s.File)]
}

// AsMap 把 Board 转成 map[Square]Piece(只含非空格)。
func (b Board) AsMap() map[Square]Piece {
	m := make(map[Square]Piece, 32)
	for r := Rank0; r <= Rank9; r++ {
		for f := FileA; f <= FileI; f++ {
			s := NewSquare(f, r)
			if p := b.PieceAt(s); p.Type != NoPieceType {
				m[s] = p
			}
		}
	}
	return m
}

// GameStatus 表示对局状态。
type GameStatus int

const (
	StatusInProgress GameStatus = iota
	StatusCheck
	StatusCheckmate
	StatusStalemate
	StatusDraw
)

func (s GameStatus) String() string {
	switch s {
	case StatusInProgress:
		return "in_progress"
	case StatusCheck:
		return "check"
	case StatusCheckmate:
		return "checkmate"
	case StatusStalemate:
		return "stalemate"
	case StatusDraw:
		return "draw"
	}
	return "unknown"
}

// DrawReason 表示和棋原因。
type DrawReason int

const (
	DrawByAgreement DrawReason = iota
	DrawBy50MoveRule
	DrawByRepetition
	DrawByStalemate
	DrawByInsufficientMaterial
)

func (d DrawReason) String() string {
	switch d {
	case DrawByAgreement:
		return "agreement"
	case DrawBy50MoveRule:
		return "fifty_move"
	case DrawByRepetition:
		return "repetition"
	case DrawByStalemate:
		return "stalemate"
	case DrawByInsufficientMaterial:
		return "insufficient_material"
	}
	return "unknown"
}

// Notation 选择着法记谱格式。
type Notation int

const (
	NotationUCI Notation = iota
	NotationChinese
	NotationChessDB
	NotationFEN
	NotationUCCI
)

func (n Notation) String() string {
	switch n {
	case NotationUCI:
		return "uci"
	case NotationChinese:
		return "chinese"
	case NotationChessDB:
		return "chessdb"
	case NotationFEN:
		return "fen"
	case NotationUCCI:
		return "ucci"
	}
	return "unknown"
}

type Move struct {
	S1   Square
	S2   Square
	Tags MoveTag
}

func (m Move) HasTag(t MoveTag) bool { return m.Tags&t != 0 }

// FileFlip 返回列镜像后的 Move。
func (m Move) FileFlip() Move {
	return Move{S1: m.S1.FileFlip(), S2: m.S2.FileFlip(), Tags: m.Tags}
}

type PGNTag struct {
	Name  string
	Value string
}

type PGNGame struct {
	Tags     []PGNTag
	Moves    []Move
	Comments []string // 与 Moves 等长,每步的评注(可为空字符串)
	Result   string
}

// Opening 表示开局库条目(ECO 编号 + 名字 + 变例)。
type Opening struct {
	Code      string
	Name      string
	Variation string
}

// GameHandle 是对局的不透明句柄,partner 通过它在 SDK 上操作有状态对局。
type GameHandle int64

// InvalidGameHandle 是无效句柄的零值。
const InvalidGameHandle GameHandle = 0

// GameOption 修改 NewGame 行为的可选参数(sealed interface)。
type GameOption interface {
	isGameOption()
}

// UseRulesOpt 启用指定规则集(如 "asian")。
type UseRulesOpt struct{ Rules string }

func (UseRulesOpt) isGameOption() {}

// IgnoreRulesOpt 禁用规则判定。
type IgnoreRulesOpt struct{}

func (IgnoreRulesOpt) isGameOption() {}

// UseNotationOpt 切换对局的记谱格式。
type UseNotationOpt struct{ N Notation }

func (UseNotationOpt) isGameOption() {}

// PieceMaterial 表示双方子力描述("RRHHC" 等)。
type PieceMaterial [2]string

// FlipDirection 棋盘翻转方向(镜像)。
type FlipDirection int

const (
	FlipUpDown    FlipDirection = iota // 上下翻转(行倒置)
	FlipLeftRight                      // 左右翻转(列镜像)
)

// Outcome 对局结果。
type Outcome string

const (
	OutcomeUnknown  Outcome = "*"       // 未结束
	OutcomeRedWon   Outcome = "1-0"     // 红方胜
	OutcomeBlackWon Outcome = "0-1"     // 黑方胜
	OutcomeDraw     Outcome = "1/2-1/2" // 和棋
)

// Verdict 亚洲规则裁决结果。
type Verdict struct {
	Outcome Outcome
	Method  Method
	Reason  string
}

// Method 终结对局的方式(与 xiangqi.Method 同步)。
type Method uint8

const (
	MethodNone Method = iota
	MethodCheckmate
	MethodResignation
	MethodDrawOffer
	MethodStalemate
	MethodThreefoldRepetition
	MethodFivefoldRepetition
	MethodFiftyMoveRule
	MethodSeventyFiveMoveRule
	// 亚洲规则附加方法
	MethodInsufficientMaterial // 官和(子力不足)
	MethodPerpetualCheck       // 长将
	MethodPerpetualChase       // 长捉
	MethodRepetition           // 允许循环/判和
	MethodRule60               // 60 回合规则(每方 120 半回合)
)

// MoveHistory 单步走子的完整记录(走子前/后位置 + 走法 + 评注)。
type MoveHistory struct {
	PreFEN   string
	PostFEN  string
	Move     Move
	Comments []string
}

// EngineHandle 是 UCI 搜索引擎进程的不透明句柄。
type EngineHandle int64

// InvalidEngineHandle 是无效句柄的零值。
const InvalidEngineHandle EngineHandle = 0

// UCICmd 是发给 UCI 引擎的命令(sealed interface)。
type UCICmd interface {
	isUCICmd()
}

// UCICmdUCINewGame 对应 "ucinewgame"。
type UCICmdUCINewGame struct{}

func (UCICmdUCINewGame) isUCICmd() {}

// UCICmdUCI 对应 "uci"(握手命令,先发这个再发 isready)。
type UCICmdUCI struct{}

func (UCICmdUCI) isUCICmd() {}

// UCICmdIsReady 对应 "isready"。
type UCICmdIsReady struct{}

func (UCICmdIsReady) isUCICmd() {}

// UCICmdSetOption 对应 "setoption name X value Y"。
type UCICmdSetOption struct {
	Name  string
	Value string
}

func (UCICmdSetOption) isUCICmd() {}

// UCICmdPosition 对应 "position [fen X|startpos] [moves ...]"。
// Moves 为空表示不发送 moves 子句。
type UCICmdPosition struct {
	StartFEN string
	Moves    []Move
}

func (UCICmdPosition) isUCICmd() {}

// UCICmdGo 对应 "go ..."。零值表示无限搜索。
type UCICmdGo struct {
	Infinite bool
	MoveTime time.Duration
	Depth    int
	Nodes    int64
	Mate     int
}

func (UCICmdGo) isUCICmd() {}

// UCICmdStop 对应 "stop"。
type UCICmdStop struct{}

func (UCICmdStop) isUCICmd() {}

// UCICmdQuit 对应 "quit"。
type UCICmdQuit struct{}

func (UCICmdQuit) isUCICmd() {}

// UCIScore 引擎评估分数。
type UCIScore struct {
	Type  string // "cp"(centipawns) 或 "mate"
	Value int
	Mate  int // >0: 几步后红胜;<0: 几步后黑胜
}

// UCIInfo 引擎搜索过程输出的一行 info。
type UCIInfo struct {
	Depth    int
	SelDepth int
	Time     int // ms
	Nodes    uint64
	Nps      int
	Score    UCIScore
	PV       []string // 主变例(走法串)
	MultiPV  int
	HashFull int
}

// UCISearchResults 一次搜索的最终结果。
type UCISearchResults struct {
	BestMove string
	Ponder   string
	Info     []UCIInfo
}

// UCIOption 引擎可配置选项。
type UCIOption struct {
	Name    string
	Type    string
	Default string
	Min     int
	Max     int
	Var     []string
}

// === Notation: UCCI (常量见上方 Notation const 块) ===

// === ChessDB 句柄 ===

type ChessDBClientHandle int64

const InvalidChessDBClientHandle ChessDBClientHandle = 0

// ChessDBCommonOptions 所有 query 的公共选项。
type ChessDBCommonOptions struct {
	Ban        []string // 禁手,如 "move:c3c4"
	EGTBMetric string   // "dtm" | "dtc"
	Learn      *bool    // nil 用服务器默认
}

type ChessDBQueryOptions struct {
	ChessDBCommonOptions
	Endgame *bool
}

type ChessDBQueryAllOptions struct {
	ChessDBCommonOptions
	ShowAll *bool
}

type ChessDBQueryRuleOptions struct {
	Reptimes int
}

// ChessDBSuggestedMove QueryBest/Query/QuerySearch 返回的候选走法。
type ChessDBSuggestedMove struct {
	Kind string // "move" | "egtb" | "search"
	Move string // 坐标走法,如 "c3c4"
}

// ChessDBMoveInfo QueryAll 返回的完整走法记录。
type ChessDBMoveInfo struct {
	Move    string
	Score   int
	Rank    int
	WinRate float64
	Note    string
}

// ChessDBScore QueryScore 返回的评估值。
type ChessDBScore struct{ Value int }

// ChessDBPV QueryPV 返回的主变例。
type ChessDBPV struct {
	Score int
	Depth int
	Moves []string
}

// ChessDBRuleResult QueryRule 返回的规则裁决。
type ChessDBRuleResult struct {
	Move string
	Rule string // "none" | "draw" | "ban"
}

// === Image 句柄 + 选项 ===

type ImageEncoderHandle int64

const InvalidImageEncoderHandle ImageEncoderHandle = 0

// ImageColor 是 image/color.Color 的简化版(避免跨 plugin 边界暴露 color 包)。
type ImageColor struct {
	R, G, B, A uint8
}

// ImageArrow 是 SVG 棋盘上从 From 到 To 的彩色箭头。
type ImageArrow struct {
	From  Square
	To    Square
	Color ImageColor
}

// ImageOption 修改 ImageSVG 行为(sealed interface)。
type ImageOption interface {
	isImageOption()
}

// ImagePerspectiveOpt 切换视角(红/黑)。
type ImagePerspectiveOpt struct{ Perspective Color }

func (ImagePerspectiveOpt) isImageOption() {}

// ImageSquareColorsOpt 自定义亮/暗格颜色。
type ImageSquareColorsOpt struct {
	Light, Dark ImageColor
}

func (ImageSquareColorsOpt) isImageOption() {}

// ImageMarkSquaresOpt 用指定颜色高亮若干格子。
type ImageMarkSquaresOpt struct {
	Color   ImageColor
	Squares []Square
}

func (ImageMarkSquaresOpt) isImageOption() {}

// ImageMarkArrowsOpt 在 SVG 上叠加若干箭头。
type ImageMarkArrowsOpt struct {
	Arrows []ImageArrow
}

func (ImageMarkArrowsOpt) isImageOption() {}
