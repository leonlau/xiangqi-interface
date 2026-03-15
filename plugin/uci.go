package main

import (
	"fmt"
	"sync"
	"testing"

	"github.com/leonlau/xiangqi"
	xq "github.com/leonlau/xiangqi-interface"
	"github.com/leonlau/xiangqi/uci"
)

// === UCI engine handle 管理 ===

var (
	engineMu      sync.Mutex
	engineCounter int64
	engines       = map[xq.EngineHandle]*uci.Engine{}
)

func getEngine(h xq.EngineHandle) (*uci.Engine, error) {
	engineMu.Lock()
	defer engineMu.Unlock()
	e, ok := engines[h]
	if !ok {
		return nil, fmt.Errorf("invalid engine handle %d", h)
	}
	return e, nil
}

// === UCI 命令转换 ===

func convertCmd(cmd xq.UCICmd) (uci.Cmd, error) {
	switch v := cmd.(type) {
	case xq.UCICmdUCI:
		return uci.CmdUCI, nil
	case xq.UCICmdUCINewGame:
		return uci.CmdUCINewGame, nil
	case xq.UCICmdIsReady:
		return uci.CmdIsReady, nil
	case xq.UCICmdStop:
		return uci.CmdStop, nil
	case xq.UCICmdQuit:
		return uci.CmdQuit, nil
	case xq.UCICmdSetOption:
		return uci.CmdSetOption{Name: v.Name, Value: v.Value}, nil
	case xq.UCICmdPosition:
		pos, err := parsePosition(v.StartFEN)
		if err != nil {
			return nil, fmt.Errorf("UCICmdPosition: invalid FEN: %w", err)
		}
		// 没有 startFEN 默认用 xiangqi.StartingPosition()
		if v.StartFEN == "" {
			pos = xiangqi.StartingPosition()
		}
		moves := make([]*xiangqi.Move, 0, len(v.Moves))
		for _, m := range v.Moves {
			internal, err := xiangqi.UCINotation{}.Decode(pos, formatUCI(m.S1)+formatUCI(m.S2))
			if err != nil {
				return nil, fmt.Errorf("UCICmdPosition: invalid move: %w", err)
			}
			moves = append(moves, internal)
			pos = pos.Update(internal) // 让下一步的合法性校验有正确上下文
		}
		return uci.CmdPosition{Position: pos, Moves: moves}, nil
	case xq.UCICmdGo:
		return uci.CmdGo{
			Infinite: v.Infinite,
			MoveTime: v.MoveTime,
			Depth:    v.Depth,
			Nodes:    int(v.Nodes),
			Mate:     v.Mate,
		}, nil
	default:
		return nil, fmt.Errorf("unknown UCICmd %T", cmd)
	}
}

// === Engine 方法实现 ===

func (XiangqiEngineImpl) NewEngine(path string) (xq.EngineHandle, error) {
	e, err := uci.New(path)
	if err != nil {
		return xq.InvalidEngineHandle, fmt.Errorf("new uci engine %q: %w", path, err)
	}
	h := xq.EngineHandle(engineCounter + 1)
	engineCounter++
	engineMu.Lock()
	engines[h] = e
	engineMu.Unlock()
	return h, nil
}

func (XiangqiEngineImpl) EngineRun(h xq.EngineHandle, cmd xq.UCICmd) error {
	e, err := getEngine(h)
	if err != nil {
		return err
	}
	xiangqiCmd, err := convertCmd(cmd)
	if err != nil {
		return err
	}
	return e.Run(xiangqiCmd)
}

func (XiangqiEngineImpl) EngineStop(h xq.EngineHandle) error {
	e, err := getEngine(h)
	if err != nil {
		return err
	}
	return e.Run(uci.CmdStop)
}

func (XiangqiEngineImpl) EngineClose(h xq.EngineHandle) error {
	engineMu.Lock()
	e, ok := engines[h]
	if ok {
		delete(engines, h)
	}
	engineMu.Unlock()
	if !ok {
		return fmt.Errorf("invalid engine handle %d", h)
	}
	return e.Close()
}

func (XiangqiEngineImpl) EngineID(h xq.EngineHandle) (map[string]string, error) {
	e, err := getEngine(h)
	if err != nil {
		return nil, err
	}
	return e.ID(), nil
}

func convertOption(o uci.Option) xq.UCIOption {
	return xq.UCIOption{
		Name:    o.Name,
		Type:    string(o.Type),
		Default: o.Default,
		Min:     atoiOrZero(o.Min),
		Max:     atoiOrZero(o.Max),
		Var:     o.Vars,
	}
}

func atoiOrZero(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}

func (XiangqiEngineImpl) EngineOptions(h xq.EngineHandle) (map[string]xq.UCIOption, error) {
	e, err := getEngine(h)
	if err != nil {
		return nil, err
	}
	src := e.Options()
	out := make(map[string]xq.UCIOption, len(src))
	for k, v := range src {
		out[k] = convertOption(v)
	}
	return out, nil
}

func convertInfo(in *uci.Info) xq.UCIInfo {
	pv := make([]string, 0, len(in.PV))
	for _, m := range in.PV {
		pv = append(pv, m.String())
	}
	score := xq.UCIScore{
		Type:  "cp",
		Value: in.Score.CP,
		Mate:  in.Score.Mate,
	}
	if in.Score.Mate != 0 {
		score.Type = "mate"
	}
	return xq.UCIInfo{
		Depth:    in.Depth,
		SelDepth: in.Seldepth,
		Time:     int(in.Time.Milliseconds()),
		Nodes:    uint64(in.Nodes),
		Nps:      in.NPS,
		Score:    score,
		PV:       pv,
		MultiPV:  in.Multipv,
		HashFull: in.Hashfull,
	}
}

func (XiangqiEngineImpl) EngineSearchResults(h xq.EngineHandle) (xq.UCISearchResults, error) {
	e, err := getEngine(h)
	if err != nil {
		return xq.UCISearchResults{}, err
	}
	src := e.SearchResults()
	out := xq.UCISearchResults{
		BestMove: "",
		Ponder:   "",
		Info:     make([]xq.UCIInfo, 0, len(src.Infos)),
	}
	if src.BestMove != nil {
		out.BestMove = src.BestMove.String()
	}
	if src.Ponder != nil {
		out.Ponder = src.Ponder.String()
	}
	for i := range src.Infos {
		out.Info = append(out.Info, convertInfo(&src.Infos[i]))
	}
	return out, nil
}

// === EngineKill ===

func (XiangqiEngineImpl) EngineKill(h xq.EngineHandle) error {
	engineMu.Lock()
	e, ok := engines[h]
	if ok {
		delete(engines, h)
	}
	engineMu.Unlock()
	if !ok {
		return fmt.Errorf("invalid engine handle %d", h)
	}
	return e.Kill()
}

// === EngineKill 测试 ===

func TestEngineKill(t *testing.T) {
	impl := &XiangqiEngineImpl{}
	h, err := impl.NewEngine("/bin/cat")
	if err != nil {
		t.Skipf("no /bin/cat on this platform: %v", err)
	}
	if err := impl.EngineKill(h); err != nil {
		t.Fatalf("EngineKill: %v", err)
	}
	// 二次调用应返回 invalid handle
	if err := impl.EngineKill(h); err == nil {
		t.Fatalf("expected error on double kill, got nil")
	}
}

func TestUCICmdGoMateForwarded(t *testing.T) {
	cmd := xq.UCICmdGo{Mate: 7}
	converted, err := convertCmd(cmd)
	if err != nil {
		t.Fatalf("convertCmd: %v", err)
	}
	got, ok := converted.(uci.CmdGo)
	if !ok {
		t.Fatalf("expected uci.CmdGo, got %T", converted)
	}
	if got.Mate != 7 {
		t.Fatalf("Mate: got %d want 7", got.Mate)
	}
}
